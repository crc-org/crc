package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// DefaultMaxResponseBytes caps successful response bodies at 32 MiB.
	DefaultMaxResponseBytes int64 = 32 << 20
	// DefaultMaxErrorBodyBytes caps error bodies at 64 KiB.
	DefaultMaxErrorBodyBytes int64 = 64 << 10
)

// ReadLimited reads at most limit bytes from r and returns an error when the
// body is larger. It prevents unbounded memory growth from misbehaving servers.
func ReadLimited(r io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		limit = DefaultMaxResponseBytes
	}
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return b, nil
}

// ReadErrorBody reads up to limit bytes from r, discarding any excess. It
// always returns a string (empty on read error) suitable for error messages.
func ReadErrorBody(r io.Reader, limit int64) string {
	if limit <= 0 {
		limit = DefaultMaxErrorBodyBytes
	}
	b, _ := io.ReadAll(io.LimitReader(r, limit)) //nolint:errcheck // best-effort error text
	return string(b)
}

// DecodeJSON decodes a single JSON value from r.
func DecodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

// HandleResponse closes resp.Body, routes non-2xx statuses through CheckStatus,
// and unmarshals the body into v when v is non-nil. A 204 No Content response
// returns without reading the body.
func HandleResponse(resp *http.Response, v any, maxBytes int64) error {
	if resp == nil {
		return &NetworkError{Cause: fmt.Errorf("nil response")}
	}
	defer resp.Body.Close() //nolint:errcheck // nothing actionable after the body is read
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusResetContent {
		return nil
	}
	body, err := ReadLimited(resp.Body, maxBytes)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if err := CheckStatus(resp, body); err != nil {
		return err
	}
	if v != nil && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("failed to decode response body: %w", err)
		}
	}
	return nil
}
