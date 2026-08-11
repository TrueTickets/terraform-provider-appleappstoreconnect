// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoPaginated(t *testing.T) {
	var server *httptest.Server

	// Two pages: page 1 links to page 2 via an absolute next URL, page 2 has
	// no next link.
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Query().Get("cursor") {
		case "":
			// First page.
			resp := map[string]any{
				"data": []map[string]any{
					{"type": "territories", "id": "USA"},
					{"type": "territories", "id": "FRA"},
				},
				"links": map[string]any{
					"next": server.URL + "/v1/territories?cursor=PAGE2&limit=200",
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "PAGE2":
			// Last page, no next link.
			resp := map[string]any{
				"data": []map[string]any{
					{"type": "territories", "id": "GBR"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client, err := NewClient("test-issuer", "test-key", testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL

	elems, err := doPaginated(context.Background(), client, Request{
		Method:   http.MethodGet,
		Endpoint: "/v1/territories",
		Query:    map[string]string{"limit": "200"},
	})
	if err != nil {
		t.Fatalf("doPaginated returned error: %v", err)
	}

	if len(elems) != 3 {
		t.Fatalf("Expected 3 elements across pages, got %d", len(elems))
	}

	var ids []string
	for _, e := range elems {
		var item struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(e, &item); err != nil {
			t.Fatalf("Failed to unmarshal element: %v", err)
		}
		ids = append(ids, item.ID)
	}

	want := []string{"USA", "FRA", "GBR"}
	for i, id := range want {
		if ids[i] != id {
			t.Errorf("element %d: expected %q, got %q", i, id, ids[i])
		}
	}
}

func TestDoPaginated_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"type": "territories", "id": "USA"},
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-issuer", "test-key", testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL

	elems, err := doPaginated(context.Background(), client, Request{
		Method:   http.MethodGet,
		Endpoint: "/v1/territories",
	})
	if err != nil {
		t.Fatalf("doPaginated returned error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("Expected 1 element, got %d", len(elems))
	}
}
