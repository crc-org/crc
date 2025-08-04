package lib

import (
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
)

// mockHTTPClient implements HTTPClient interface for testing
type mockHTTPClient struct {
	responses map[string]*http.Response
	errors    map[string]error
	requests  []*http.Request
	mu        sync.RWMutex
}

func newMockHTTPClient() *mockHTTPClient {
	return &mockHTTPClient{
		responses: make(map[string]*http.Response),
		errors:    make(map[string]error),
		requests:  make([]*http.Request, 0),
	}
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = append(m.requests, req)

	key := req.Method + ":" + req.URL.String()

	if err, exists := m.errors[key]; exists {
		return nil, err
	}

	if resp, exists := m.responses[key]; exists {
		return resp, nil
	}

	// Default response
	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		Body:          io.NopCloser(strings.NewReader("test content")),
		ContentLength: 12,
		Header:        make(http.Header),
		Request:       req,
	}, nil
}

func (m *mockHTTPClient) addResponse(method, url string, resp *http.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[method+":"+url] = resp
}

func (m *mockHTTPClient) addError(method, url string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[method+":"+url] = err
}

func (m *mockHTTPClient) getRequests() []*http.Request {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]*http.Request(nil), m.requests...)
}

// testDirectoryManager manages temporary directory creation and cleanup for tests
type testDirectoryManager struct {
	tempDir     string
	originalDir string
}

// setupTestDirectory creates a temporary directory and changes to it
func setupTestDirectory(t *testing.T, prefix string) *testDirectoryManager {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	return &testDirectoryManager{
		tempDir:     tempDir,
		originalDir: originalDir,
	}
}

// cleanup restores the original directory and removes the temporary directory
func (tdm *testDirectoryManager) cleanup() {
	_ = os.Chdir(tdm.originalDir)
	_ = os.RemoveAll(tdm.tempDir)
}

// setupTestDirectoryWithCleanup sets up a test directory and automatically schedules cleanup
func setupTestDirectoryWithCleanup(t *testing.T, prefix string) {
	t.Helper()
	tdm := setupTestDirectory(t, prefix)
	t.Cleanup(tdm.cleanup)
}

// createMockHTTPResponse creates a standard mock HTTP response
func createMockHTTPResponse(status string, statusCode int, content string, headers map[string]string) *http.Response {
	resp := &http.Response{
		Status:        status,
		StatusCode:    statusCode,
		Proto:         "HTTP/1.1",
		Body:          io.NopCloser(strings.NewReader(content)),
		ContentLength: int64(len(content)),
		Header:        make(http.Header),
	}

	for key, value := range headers {
		resp.Header.Set(key, value)
	}

	return resp
}

// createSuccessResponse creates a standard 200 OK response
func createSuccessResponse(content string) *http.Response {
	return createMockHTTPResponse("200 OK", 200, content, nil)
}

// createErrorResponse creates a standard error response
func createErrorResponse(statusCode int, content string) *http.Response {
	status := http.StatusText(statusCode)
	return createMockHTTPResponse(status, statusCode, content, nil)
}

// withMockClient temporarily replaces the default client and restores it
func withMockClient(t *testing.T, mockClient *mockHTTPClient, fn func()) {
	t.Helper()

	originalClient := DefaultClient
	DefaultClient = &Client{
		HTTPClient: mockClient,
		UserAgent:  "test-agent",
	}

	defer func() {
		DefaultClient = originalClient
	}()

	fn()
}

// createTestClient creates a client for testing with the given HTTP client
