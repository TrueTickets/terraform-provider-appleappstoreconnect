// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testPrivateKey = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgbnVpCD6pN8HEhE7F
RvnHE7a8kGGAC2vq3fKj6TIJiY+hRANCAATgNg3pjdLU5J7J9hxFyHHxvRnB3v5y
DC4N2IwV3YFHSBnFq72u5nkFmPKJiCLSLBDFkWtQPGBXpUDNhD9L3kOh
-----END PRIVATE KEY-----`

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		issuerID   string
		keyID      string
		privateKey string
		wantErr    bool
	}{
		{
			name:       "valid client",
			issuerID:   "test-issuer",
			keyID:      "test-key",
			privateKey: testPrivateKey,
			wantErr:    false,
		},
		{
			name:       "invalid private key",
			issuerID:   "test-issuer",
			keyID:      "test-key",
			privateKey: "invalid-key",
			wantErr:    true,
		},
		{
			name:       "empty issuer ID",
			issuerID:   "",
			keyID:      "test-key",
			privateKey: testPrivateKey,
			wantErr:    true,
		},
		{
			name:       "empty key ID",
			issuerID:   "test-issuer",
			keyID:      "",
			privateKey: testPrivateKey,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.issuerID, tt.keyID, tt.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_GetToken(t *testing.T) {
	client, err := NewClient("test-issuer", "test-key", testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test initial token generation
	token1, err := client.getToken()
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if token1 == "" {
		t.Error("Generated token is empty")
	}

	// Test token caching
	token2, err := client.getToken()
	if err != nil {
		t.Fatalf("Failed to get second token: %v", err)
	}

	if token1 != token2 {
		t.Error("Expected cached token, got new token")
	}
}

func TestClient_Do(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate authorization header
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Invalid authorization header: %s", authHeader)
		}

		// Validate content type
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Invalid content type: %s", contentType)
			}
		}

		// Return response based on path
		switch r.URL.Path {
		case "/v1/passTypeIds":
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"type": "passTypeIds",
					"id":   "test-id",
					"attributes": map[string]interface{}{
						"passTypeIdentifier": "pass.com.example.test",
						"name":               "Test Pass",
						"description":        "Test Description",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		case "/v1/error":
			w.WriteHeader(http.StatusBadRequest)
			errorResp := map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"status": "400",
						"code":   "INVALID_REQUEST",
						"title":  "Invalid Request",
						"detail": "Test error",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(errorResp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client, err := NewClient("test-issuer", "test-key", testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Override the baseURL for testing
	client.baseURL = server.URL + "/v1"

	ctx := context.Background()

	// Test successful request
	t.Run("successful request", func(t *testing.T) {
		req := Request{
			Method:   http.MethodGet,
			Endpoint: "/passTypeIds",
		}

		resp, err := client.Do(ctx, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Response is nil")
		}

		// Check that we got data back
		if resp.Data == nil {
			t.Fatal("Response data is nil")
		}

		// The resp.Data should contain the JSON response
		// Let's just verify it's not empty
		if len(resp.Data) == 0 {
			t.Fatal("Response data is empty")
		}
	})

	// Test error response
	t.Run("error response", func(t *testing.T) {
		req := Request{
			Method:   http.MethodGet,
			Endpoint: "/error",
		}

		_, err := client.Do(ctx, req)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if !strings.Contains(err.Error(), "Test error") {
			t.Errorf("Error should contain 'Test error', got: %v", err)
		}
	})

	// Test with query parameters
	t.Run("with query parameters", func(t *testing.T) {
		req := Request{
			Method:   http.MethodGet,
			Endpoint: "/passTypeIds",
			Query: map[string]string{
				"filter[identifier]": "pass.com.example.test",
				"include":            "certificates",
			},
		}

		resp, err := client.Do(ctx, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Response is nil")
		}
	})

	// Test with body
	t.Run("with body", func(t *testing.T) {
		body := map[string]interface{}{
			"data": map[string]interface{}{
				"type": "passTypeIds",
				"attributes": map[string]interface{}{
					"passTypeIdentifier": "pass.com.example.new",
					"name":               "New Pass",
				},
			},
		}

		bodyBytes, _ := json.Marshal(body)

		req := Request{
			Method:   http.MethodPost,
			Endpoint: "/passTypeIds",
			Body:     bodyBytes,
		}

		resp, err := client.Do(ctx, req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp == nil {
			t.Fatal("Response is nil")
		}
	})
}

func TestClient_TokenExpiration(t *testing.T) {
	client, err := NewClient("test-issuer", "test-key", testPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Get initial token
	token1, err := client.getToken()
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	// Manually expire the token
	client.mu.Lock()
	client.tokenExpiry = time.Now().Add(-1 * time.Hour)
	client.mu.Unlock()

	// Get new token (should be different due to expiration)
	token2, err := client.getToken()
	if err != nil {
		t.Fatalf("Failed to get new token: %v", err)
	}

	if token1 == token2 {
		t.Error("Expected new token after expiration, got cached token")
	}
}
