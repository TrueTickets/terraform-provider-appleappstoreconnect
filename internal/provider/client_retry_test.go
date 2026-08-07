// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newRetryTestClient returns a client pointed at server with a negligible
// backoff so the retry tests stay fast.
func newRetryTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()

	client, err := NewClient("test-issuer", "test-key", testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	client.baseURL = serverURL
	client.retryBackoff = time.Millisecond

	return client
}

// TestDoRetriesServerErrors covers the transient 500 UNEXPECTED_ERROR App Store
// Connect returns on writes -- notably when several localizations of the same
// parent are created concurrently. Without the retry the apply fails and the
// operator has to rerun it by hand.
func TestDoRetriesServerErrors(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"errors":[{"status":"500","code":"UNEXPECTED_ERROR","title":"An unexpected error occurred."}]}`))

			return
		}

		_, _ = w.Write([]byte(`{"data":{"id":"CREATED"}}`))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)

	resp, err := client.Do(context.Background(), Request{
		Method:   http.MethodPost,
		Endpoint: "/passTypeIds",
		Body:     map[string]any{"identifier": "pass.io.truetickets.test"},
	})
	if err != nil {
		t.Fatalf("Expected the third attempt to succeed, got error: %v", err)
	}

	if got := calls.Load(); got != 3 {
		t.Errorf("Expected 3 attempts, got %d", got)
	}

	if string(resp.Data) != `{"id":"CREATED"}` {
		t.Errorf("Unexpected payload: %s", resp.Data)
	}
}

// TestDoGivesUpAfterMaxAttempts pins the bound: a server that never recovers
// must surface the error instead of looping.
func TestDoGivesUpAfterMaxAttempts(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)

	if _, err := client.Do(context.Background(), Request{
		Method:   http.MethodGet,
		Endpoint: "/passTypeIds",
	}); err == nil {
		t.Fatal("Expected an error once the attempts are exhausted")
	}

	if got := calls.Load(); got != int32(maxRetryAttempts) {
		t.Errorf("Expected %d attempts, got %d", maxRetryAttempts, got)
	}
}

// TestDoDoesNotRetryClientErrors guards the other half of the policy: a 4xx is
// a deterministic rejection, replaying it only repeats the same answer.
func TestDoDoesNotRetryClientErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
	}{
		{name: "not found", status: http.StatusNotFound},
		{name: "conflict", status: http.StatusConflict},
		{name: "unprocessable", status: http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			client := newRetryTestClient(t, server.URL)

			if _, err := client.Do(context.Background(), Request{
				Method:   http.MethodDelete,
				Endpoint: "/certificates/CERT_ID",
			}); err == nil {
				t.Fatal("Expected an error")
			}

			if got := calls.Load(); got != 1 {
				t.Errorf("Expected a single attempt, got %d", got)
			}
		})
	}
}

// TestDoRetryReplaysTheBody makes sure every attempt carries the payload: the
// body reader is consumed by the first attempt, so it has to be rebuilt.
func TestDoRetryReplaysTheBody(t *testing.T) {
	t.Parallel()

	var (
		calls  atomic.Int32
		bodies = make(chan string, 3)
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		bodies <- string(buf)

		w.Header().Set("Content-Type", "application/json")

		if calls.Add(1) < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		_, _ = w.Write([]byte(`{"data":{"id":"OK"}}`))
	}))
	defer server.Close()

	client := newRetryTestClient(t, server.URL)

	if _, err := client.Do(context.Background(), Request{
		Method:   http.MethodPost,
		Endpoint: "/passTypeIds",
		Body:     map[string]any{"identifier": "pass.io.truetickets.test"},
	}); err != nil {
		t.Fatalf("Expected the retry to succeed, got error: %v", err)
	}

	close(bodies)

	seen := 0

	for body := range bodies {
		seen++

		if body != `{"identifier":"pass.io.truetickets.test"}` {
			t.Errorf("attempt %d sent %q", seen, body)
		}
	}

	if seen != 2 {
		t.Errorf("Expected 2 bodies, got %d", seen)
	}
}
