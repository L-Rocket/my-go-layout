// Package postgrest implements the domain Store interfaces against a
// Supabase PostgREST API. main wires these only when SUPABASE_URL and
// SUPABASE_SERVICE_ROLE_KEY are both set.
package postgrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the shared PostgREST client: base URL, service-role auth
// headers, and JSON encode/decode. Store implementations embed it.
type Client struct {
	baseURL    string
	serviceKey string
	http       *http.Client
}

// New builds a client for <supabaseURL>/rest/v1 using the service-role
// key. The key must never reach browsers — it bypasses RLS.
func New(supabaseURL, serviceKey string) *Client {
	return &Client{
		baseURL:    supabaseURL + "/rest/v1",
		serviceKey: serviceKey,
		http:       &http.Client{Timeout: 10 * time.Second},
	}
}

// do runs one PostgREST request. body is JSON-encoded when non-nil; out
// is JSON-decoded when non-nil. prefer sets the PostgREST Prefer header
// (e.g. "return=representation" so inserts/patches return the row).
func (c *Client) do(ctx context.Context, method, path string, body, out any, prefer string) error {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("postgrest: encode body: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return fmt.Errorf("postgrest: build request: %w", err)
	}
	// PostgREST authenticates with BOTH headers carrying the service key.
	req.Header.Set("apikey", c.serviceKey)
	req.Header.Set("Authorization", "Bearer "+c.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	if prefer != "" {
		req.Header.Set("Prefer", prefer)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("postgrest: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("postgrest: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{Status: resp.StatusCode, Body: string(data)}
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("postgrest: decode response: %w", err)
		}
	}
	return nil
}

// HTTPError is a non-2xx PostgREST response, kept structured so stores
// can map statuses (e.g. 409 → duplicate) to domain errors.
type HTTPError struct {
	Status int
	Body   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("postgrest: HTTP %d: %s", e.Status, e.Body)
}
