package httpx

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultTimeout caps a whole request.
	DefaultTimeout = 30 * time.Second
	// DefaultDialTimeout caps connection establishment.
	DefaultDialTimeout = 10 * time.Second
	// DefaultTLSHandshakeTimeout caps the TLS handshake.
	DefaultTLSHandshakeTimeout = 10 * time.Second
)

// Options configures a Client. The zero value is usable: it produces a client
// with the package defaults, no base URL, and no extra headers.
type Options struct {
	// BaseURL is prepended to relative request paths.
	BaseURL string
	// Timeout is the overall request timeout. Defaults to DefaultTimeout.
	Timeout time.Duration
	// DialTimeout is the connection timeout. Defaults to DefaultDialTimeout.
	DialTimeout time.Duration
	// TLSHandshakeTimeout defaults to DefaultTLSHandshakeTimeout.
	TLSHandshakeTimeout time.Duration
	// Transport replaces the default transport entirely when non-nil.
	Transport http.RoundTripper
	// TLSConfig is applied to the default transport when non-nil.
	TLSConfig *tls.Config
	// UserAgent, when set, is sent on every request.
	UserAgent string
	// Headers are added to every request (Authorization, Accept, ...).
	Headers http.Header
	// MaxResponseBytes caps successful bodies. Defaults to DefaultMaxResponseBytes.
	MaxResponseBytes int64
	// MaxErrorBodyBytes caps error bodies. Defaults to DefaultMaxErrorBodyBytes.
	MaxErrorBodyBytes int64
	// Logger, when set, receives debug logs for each request.
	Logger *slog.Logger
}

// Client is a thin wrapper around *http.Client with a base URL, default
// headers, bounded body reads and typed error handling.
type Client struct {
	BaseURL           string
	HTTPClient        *http.Client
	UserAgent         string
	Headers           http.Header
	MaxResponseBytes  int64
	MaxErrorBodyBytes int64
	Logger            *slog.Logger

	// SetHeaders, when non-nil, runs on every request after the client's
	// static headers are applied. Use it for per-request authentication.
	SetHeaders func(*http.Request)
}

// NewClient builds a Client from opts, filling in defaults.
func NewClient(opts Options) *Client { //nolint:gocritic // Options is a value type copied once at construction
	return &Client{
		BaseURL:           opts.BaseURL,
		HTTPClient:        NewHTTPClient(opts),
		UserAgent:         opts.UserAgent,
		Headers:           opts.Headers.Clone(),
		MaxResponseBytes:  opts.MaxResponseBytes,
		MaxErrorBodyBytes: opts.MaxErrorBodyBytes,
		Logger:            opts.Logger,
	}
}

// NewHTTPClient builds an *http.Client with the configured timeouts and
// transport. The default transport is cloned so proxy settings and HTTP/2
// support are preserved.
func NewHTTPClient(opts Options) *http.Client { //nolint:gocritic // Options is a value type copied once at construction
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	dialTimeout := opts.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = DefaultDialTimeout
	}
	tlsTimeout := opts.TLSHandshakeTimeout
	if tlsTimeout <= 0 {
		tlsTimeout = DefaultTLSHandshakeTimeout
	}

	transport := opts.Transport
	if transport == nil {
		base, ok := http.DefaultTransport.(*http.Transport)
		if ok {
			transport = base.Clone()
		} else {
			transport = &http.Transport{}
		}
		tr, ok := transport.(*http.Transport)
		if ok {
			tr.DialContext = (&net.Dialer{Timeout: dialTimeout}).DialContext
			tr.TLSHandshakeTimeout = tlsTimeout
			if opts.TLSConfig != nil {
				tr.TLSClientConfig = opts.TLSConfig
			}
		}
	}

	return &http.Client{Timeout: timeout, Transport: transport}
}

// URL resolves path against the client's base URL. Absolute URLs are returned
// unchanged. Path elements are not escaped: callers appending user-supplied
// IDs should wrap them in PathSeg.
func (c *Client) URL(path string) string {
	if path == "" {
		return c.BaseURL
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if c.BaseURL == "" {
		return path
	}
	return JoinPath(c.BaseURL, path)
}

// NewRequest builds a request for an API path (or an absolute URL). A non-nil
// body is marshalled as JSON.
func (c *Client) NewRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	return NewRequestWithBody(ctx, method, c.URL(path), body)
}

// Do applies the client's headers and executes req.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.applyHeaders(req)
	if c.Logger != nil {
		c.Logger.Debug("http request",
			slog.String("method", req.Method),
			slog.String("url", RedactURLQuery(req.URL)),
		)
	}
	resp, err := c.HTTPClient.Do(req)
	if c.Logger != nil && resp != nil {
		c.Logger.Debug("http response",
			slog.String("method", req.Method),
			slog.String("url", RedactURLQuery(req.URL)),
			slog.Int("status", resp.StatusCode),
		)
	}
	return resp, err
}

// DoJSON builds a JSON request, executes it, and decodes a successful response
// into out. A nil out discards the body. Transport failures are wrapped in
// *NetworkError; non-2xx responses are returned as typed errors.
func (c *Client) DoJSON(ctx context.Context, method, path string, body, out any) error {
	req, err := c.NewRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return AsNetworkError(err)
	}
	return HandleResponse(resp, out, c.MaxResponseBytes)
}

func (c *Client) applyHeaders(req *http.Request) {
	for key, values := range c.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if c.UserAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.SetHeaders != nil {
		c.SetHeaders(req)
	}
}
