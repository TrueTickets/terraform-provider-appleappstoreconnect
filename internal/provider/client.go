// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// baseURL is the base URL for the App Store Connect API. It intentionally
	// omits the API version: each request's Endpoint carries its own version
	// prefix (e.g. "/v1/passTypeIds", "/v2/inAppPurchases") because the API
	// mixes v1 and v2 endpoints.
	baseURL = "https://api.appstoreconnect.apple.com"

	// maxRetryAttempts bounds how many times a request answered with a 5xx is
	// replayed, and retryBackoff is the base delay between attempts (it grows
	// linearly: 1s, then 2s).
	maxRetryAttempts = 3
	retryBackoff     = time.Second

	// tokenExpiration is the maximum lifetime of a JWT token (20 minutes).
	tokenExpiration = 20 * time.Minute

	// tokenRefreshBuffer is the buffer time before token expiration to refresh.
	tokenRefreshBuffer = 5 * time.Minute
)

// Client represents an App Store Connect API client.
type Client struct {
	httpClient *http.Client
	issuerID   string
	keyID      string
	privateKey interface{}
	baseURL    string

	// Retry policy for transient 5xx answers. Fields rather than constants so
	// tests can shrink the backoff.
	maxAttempts  int
	retryBackoff time.Duration

	// Token management
	mu           sync.RWMutex
	currentToken string
	tokenExpiry  time.Time
}

// NewClient creates a new App Store Connect API client.
func NewClient(issuerID, keyID, privateKeyPEM string) (*Client, error) {
	// Validate inputs
	if issuerID == "" {
		return nil, fmt.Errorf("issuer ID cannot be empty")
	}
	if keyID == "" {
		return nil, fmt.Errorf("key ID cannot be empty")
	}
	if privateKeyPEM == "" {
		return nil, fmt.Errorf("private key cannot be empty")
	}

	// Parse the private key from PEM format
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse private key PEM block")
	}

	// Parse the key based on the type
	var privateKey interface{}
	var err error

	switch block.Type {
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		issuerID:   issuerID,
		keyID:      keyID,
		privateKey: privateKey,
		baseURL:    baseURL,

		maxAttempts:  maxRetryAttempts,
		retryBackoff: retryBackoff,
	}, nil
}

// generateToken generates a new JWT token for API authentication.
func (c *Client) generateToken() (string, error) {
	now := time.Now()

	// Create the claims
	claims := jwt.MapClaims{
		"iss": c.issuerID,
		"iat": now.Unix(),
		"exp": now.Add(tokenExpiration).Unix(),
		"aud": "appstoreconnect-v1",
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID

	// Sign the token
	tokenString, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// getToken returns a valid token, refreshing if necessary.
func (c *Client) getToken() (string, error) {
	c.mu.RLock()
	if c.currentToken != "" && time.Now().Before(c.tokenExpiry.Add(-tokenRefreshBuffer)) {
		token := c.currentToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	// Need to refresh token
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.currentToken != "" && time.Now().Before(c.tokenExpiry.Add(-tokenRefreshBuffer)) {
		return c.currentToken, nil
	}

	// Generate new token
	token, err := c.generateToken()
	if err != nil {
		return "", err
	}

	c.currentToken = token
	c.tokenExpiry = time.Now().Add(tokenExpiration)

	return token, nil
}

// Request represents a generic API request.
type Request struct {
	Method   string
	Endpoint string
	Body     interface{}
	Query    map[string]string
}

// Response represents a generic API response.
type Response struct {
	Data     json.RawMessage `json:"data"`
	Errors   []Error         `json:"errors,omitempty"`
	Links    Links           `json:"links,omitempty"`
	Meta     Meta            `json:"meta,omitempty"`
	Included json.RawMessage `json:"included,omitempty"`
}

// Error represents an API error.
type Error struct {
	ID     string       `json:"id,omitempty"`
	Status string       `json:"status,omitempty"`
	Code   string       `json:"code,omitempty"`
	Title  string       `json:"title,omitempty"`
	Detail string       `json:"detail,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
}

// APIError is returned by Client.Do for any non-2xx HTTP response. Callers can
// inspect StatusCode to react to specific conditions (e.g. treat 404 as a
// drift signal and remove the resource from state) via errors.As.
type APIError struct {
	StatusCode int
	Errors     []Error
	RawBody    string
}

// Error formats the API error. The format is preserved from the previous
// inline fmt.Errorf calls so Diagnostics messages stay stable.
func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		parts := make([]string, 0, len(e.Errors))
		for _, apiErr := range e.Errors {
			parts = append(parts, fmt.Sprintf("%s: %s", apiErr.Title, apiErr.Detail))
		}
		return "API error: " + strings.Join(parts, "; ")
	}
	if e.RawBody != "" {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.RawBody)
	}
	return fmt.Sprintf("API error: HTTP %d", e.StatusCode)
}

// IsNotFound reports whether err is an *APIError with HTTP 404 status.
func IsNotFound(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusNotFound
}

// ErrorSource represents the source of an error.
type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

// Links represents pagination links.
type Links struct {
	Self  string `json:"self,omitempty"`
	First string `json:"first,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Last  string `json:"last,omitempty"`
}

// Meta represents response metadata.
type Meta struct {
	Paging *Paging `json:"paging,omitempty"`
}

// Paging represents pagination metadata.
type Paging struct {
	Total int `json:"total"`
	Limit int `json:"limit"`
}

// Do performs the request, retrying transient server-side failures.
//
// App Store Connect intermittently answers 500 UNEXPECTED_ERROR on writes --
// most visibly when several localizations of the same parent are created
// concurrently -- and the very same request succeeds moments later. Without a
// retry every such blip surfaces as an apply error the operator has to rerun
// by hand.
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	// Marshal the body once: each attempt needs its own reader over the bytes.
	var bodyBytes []byte

	if req.Body != nil {
		var err error

		bodyBytes, err = json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		tflog.Debug(ctx, "API request body", map[string]interface{}{
			"body": string(bodyBytes),
		})
	}

	var lastErr error

	for attempt := 1; attempt <= c.maxAttempts; attempt++ {
		resp, err := c.doOnce(ctx, req, bodyBytes)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if attempt == c.maxAttempts || !isRetryable(err) {
			return nil, err
		}

		delay := c.retryBackoff * time.Duration(attempt)

		tflog.Debug(ctx, "Retrying API request after a server error", map[string]interface{}{
			"method":   req.Method,
			"endpoint": req.Endpoint,
			"attempt":  attempt,
			"delay":    delay.String(),
			"error":    err.Error(),
		})

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, lastErr
}

// isRetryable reports whether err is a transient server-side failure. Only 5xx
// qualifies: a 4xx is a deterministic rejection and retrying it just repeats
// the same answer.
func isRetryable(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}

	return apiErr.StatusCode >= http.StatusInternalServerError
}

// doOnce performs a single attempt of req.
func (c *Client) doOnce(ctx context.Context, req Request, bodyBytes []byte) (*Response, error) {
	// Build URL
	urlStr := c.baseURL + req.Endpoint

	// Add query parameters
	if len(req.Query) > 0 {
		params := url.Values{}
		for key, value := range req.Query {
			params.Add(key, value)
		}
		urlStr += "?" + params.Encode()
	}

	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get token
	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication token: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	tflog.Debug(ctx, "Making API request", map[string]interface{}{
		"method":   req.Method,
		"endpoint": req.Endpoint,
		"url":      urlStr,
	})

	// Perform request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	tflog.Debug(ctx, "API response", map[string]interface{}{
		"status": httpResp.StatusCode,
		"body":   string(respBody),
	})

	// Parse response
	var resp Response
	// Handle empty responses (common for DELETE operations)
	if len(respBody) == 0 {
		// For successful DELETE operations, return empty response
		if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
			return &resp, nil
		}
		// For error responses that are empty, return a typed APIError so
		// callers can branch on the status code (e.g. 404 → RemoveResource).
		return nil, &APIError{StatusCode: httpResp.StatusCode, RawBody: "empty response"}
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		// If we can't parse as a standard response, check if it's an error
		if httpResp.StatusCode >= 400 {
			return nil, &APIError{StatusCode: httpResp.StatusCode, RawBody: string(respBody)}
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if len(resp.Errors) > 0 {
		return nil, &APIError{StatusCode: httpResp.StatusCode, Errors: resp.Errors}
	}

	// Check HTTP status
	if httpResp.StatusCode >= 400 {
		return nil, &APIError{StatusCode: httpResp.StatusCode}
	}

	return &resp, nil
}
