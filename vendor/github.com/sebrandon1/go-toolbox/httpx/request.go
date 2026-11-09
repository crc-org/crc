package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AddQueryParams adds each key/value to the request URL's query string,
// preserving any existing parameters.
func AddQueryParams(req *http.Request, params map[string]string) {
	if len(params) == 0 {
		return
	}
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
}

// NewRequest builds a request with the given context, method and URL.
func NewRequest(ctx context.Context, method, rawURL string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, rawURL, nil)
}

// NewRequestWithBody builds a request whose body is body marshalled as JSON.
// A nil body produces a request with no body.
func NewRequestWithBody(ctx context.Context, method, rawURL string, body any) (*http.Request, error) {
	if body == nil {
		return NewRequest(ctx, method, rawURL)
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(data))
}

// PathSeg escapes a single path segment so IDs containing '/', '..' or
// percent sequences cannot rewrite the request path.
func PathSeg(s string) string {
	return url.PathEscape(s)
}

// JoinPath joins a base URL with path elements, escaping each element and
// cleaning the result. It is the safe way to append user-supplied IDs.
func JoinPath(base string, elems ...string) string {
	joined, err := url.JoinPath(base, elems...)
	if err != nil {
		// url.JoinPath only fails when an element re-parses to a different
		// path; falling back to the base keeps the failure visible upstream
		// as a request error rather than panicking here.
		return base
	}
	return joined
}

// RedactURLQuery returns a copy of rawURL as a string with credentials and
// sensitive query parameter values replaced by "[REDACTED]". It is intended
// for logging.
func RedactURLQuery(rawURL *url.URL) string {
	if rawURL == nil {
		return "<unknown>"
	}

	redacted := *rawURL
	redacted.User = nil
	query := redacted.Query()
	for key := range query {
		if isSensitiveKey(key) {
			query[key] = []string{"[REDACTED]"}
		}
	}
	redacted.RawQuery = query.Encode()
	return redacted.String()
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, marker := range []string{"token", "secret", "password", "auth", "api_key", "apikey"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// CloneBody returns a function that produces a fresh reader over body, for
// handlers that need to retry a request. It returns nil when body is nil.
func CloneBody(body []byte) func() io.Reader {
	if body == nil {
		return nil
	}
	return func() io.Reader { return bytes.NewReader(body) }
}
