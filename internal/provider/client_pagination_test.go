// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// idsOf unmarshals a JSON:API data array and returns its ids in order.
func idsOf(t *testing.T, data json.RawMessage) []string {
	t.Helper()

	var items []struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	ids := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}

	return ids
}

// TestDoListFollowsPagination is the regression guard: the App Store Connect
// API hands back one page at a time with a links.next cursor. Do reads only the
// first page (documented here), while DoList walks every page and returns the
// whole collection. Without the follow, a certificate or pass type ID on a
// later page is invisible to the provider.
func TestDoListFollowsPagination(t *testing.T) {
	t.Parallel()

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("cursor") {
		case "":
			next := server.URL + "/certificates?cursor=P2&limit=200"
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"A"}],"links":{"next":%q}}`, next)))
		case "P2":
			next := server.URL + "/certificates?cursor=P3&limit=200"
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"B"}],"links":{"next":%q}}`, next)))
		case "P3":
			_, _ = w.Write([]byte(`{"data":[{"id":"C"}]}`))
		default:
			t.Errorf("unexpected cursor %q", r.URL.Query().Get("cursor"))
		}
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)

	req := Request{
		Method:   http.MethodGet,
		Endpoint: "/certificates",
		Query:    map[string]string{"limit": "200"},
	}

	// Old single-page read only ever sees the first page -- this is the bug
	// DoList fixes.
	single, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}

	if got := idsOf(t, single.Data); len(got) != 1 || got[0] != "A" {
		t.Fatalf("Do returned %v, want just [A]", got)
	}

	// DoList walks the whole collection.
	all, err := client.DoList(context.Background(), req)
	if err != nil {
		t.Fatalf("DoList: %v", err)
	}

	got := idsOf(t, all.Data)
	want := []string{"A", "B", "C"}

	if len(got) != len(want) {
		t.Fatalf("DoList returned %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("DoList returned %v, want %v", got, want)
		}
	}

	// The accumulated Links must not leak a stale next cursor to the caller.
	if all.Links.Next != "" {
		t.Errorf("expected cleared links.next, got %q", all.Links.Next)
	}
}

// TestDoListSinglePage covers the common case: one page, no next cursor, no
// extra round-trips.
func TestDoListSinglePage(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"A"},{"id":"B"}]}`))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)

	all, err := client.DoList(context.Background(), Request{
		Method:   http.MethodGet,
		Endpoint: "/passTypeIds",
	})
	if err != nil {
		t.Fatalf("DoList: %v", err)
	}

	if got := idsOf(t, all.Data); len(got) != 2 {
		t.Fatalf("got %v, want 2 items", got)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("expected a single request, got %d", got)
	}
}

// TestDoListEmpty makes sure an empty collection round-trips as a valid empty
// JSON array rather than null, so callers unmarshalling into a slice keep
// working.
func TestDoListEmpty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)

	all, err := client.DoList(context.Background(), Request{
		Method:   http.MethodGet,
		Endpoint: "/certificates",
	})
	if err != nil {
		t.Fatalf("DoList: %v", err)
	}

	if string(all.Data) != "[]" {
		t.Errorf("expected empty array, got %s", all.Data)
	}
}

// TestDoListRetriesPageErrors confirms the retry policy that guards the first
// request also guards each follow-up page fetch: a transient 500 on page two is
// replayed rather than truncating the collection.
func TestDoListRetriesPageErrors(t *testing.T) {
	t.Parallel()

	var page2Calls atomic.Int32

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("cursor") == "P2" {
			if page2Calls.Add(1) < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, _ = w.Write([]byte(`{"data":[{"id":"B"}]}`))

			return
		}

		next := server.URL + "/certificates?cursor=P2"
		_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"A"}],"links":{"next":%q}}`, next)))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)
	client.retryBackoff = time.Millisecond

	all, err := client.DoList(context.Background(), Request{
		Method:   http.MethodGet,
		Endpoint: "/certificates",
	})
	if err != nil {
		t.Fatalf("DoList: %v", err)
	}

	if got := idsOf(t, all.Data); len(got) != 2 {
		t.Fatalf("got %v, want [A B]", got)
	}

	if got := page2Calls.Load(); got != 2 {
		t.Errorf("expected page 2 to be retried once (2 calls), got %d", got)
	}
}

// TestDoListStopsAtMaxPages guards against a server that keeps handing back a
// self-referential cursor: DoList must terminate rather than loop forever.
func TestDoListStopsAtMaxPages(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		next := server.URL + "/certificates?cursor=LOOP"
		_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"X"}],"links":{"next":%q}}`, next)))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)

	if _, err := client.DoList(context.Background(), Request{
		Method:   http.MethodGet,
		Endpoint: "/certificates",
	}); err != nil {
		t.Fatalf("DoList: %v", err)
	}

	// One initial request plus at most maxListPages follow-ups.
	if got := calls.Load(); got > int32(maxListPages)+1 {
		t.Errorf("DoList made %d requests, expected it to stop by %d", got, maxListPages+1)
	}
}
