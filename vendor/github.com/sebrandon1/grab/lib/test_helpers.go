package lib

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
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

// getFallbackURLs returns a list of fallback URLs for a given size in bytes
// These URLs provide content from different reliable sources with high availability
func getFallbackURLs(size int) []string {
	urls := []string{}

	// Fallback 1: httpbin.org (original, but moved to fallback)
	urls = append(urls, fmt.Sprintf("https://httpbin.org/bytes/%d", size))

	// Fallback 2: GitHub's own repository files (highly reliable)
	urls = append(urls, "https://github.com/sebrandon1/grab/archive/refs/heads/main.zip")

	// Fallback 3: Go language official downloads (very stable)
	urls = append(urls, "https://go.dev/dl/go1.21.5.src.tar.gz")

	// Fallback 4: Raw GitHub files based on size (specific content)
	switch size {
	case 256:
		urls = append(urls,
			"https://raw.githubusercontent.com/golang/go/master/src/go/build/testdata/empty/dummy",
			"https://raw.githubusercontent.com/golang/go/master/VERSION",
			"https://raw.githubusercontent.com/microsoft/vscode/main/.eslintrc.json")
	case 512:
		urls = append(urls,
			"https://raw.githubusercontent.com/kubernetes/kubernetes/master/.gitignore",
			"https://raw.githubusercontent.com/golang/go/master/CONTRIBUTORS",
			"https://raw.githubusercontent.com/golang/go/master/src/cmd/go/testdata/script/mod_tidy_compat.txt")
	case 768:
		urls = append(urls,
			"https://raw.githubusercontent.com/microsoft/vscode/main/package.json",
			"https://raw.githubusercontent.com/golang/go/master/src/go/doc/comment.go")
	case 1024:
		urls = append(urls,
			"https://raw.githubusercontent.com/golang/go/master/LICENSE",
			"https://raw.githubusercontent.com/kubernetes/kubernetes/master/README.md",
			"https://raw.githubusercontent.com/golang/go/master/SECURITY.md")
	case 2048:
		urls = append(urls,
			"https://raw.githubusercontent.com/kubernetes/kubernetes/master/go.mod",
			"https://raw.githubusercontent.com/golang/go/master/go.mod")
	default:
		// For other sizes, add stable large files
		urls = append(urls,
			"https://raw.githubusercontent.com/golang/go/master/README.md",
			"https://raw.githubusercontent.com/microsoft/vscode/main/package.json",
			"https://raw.githubusercontent.com/golang/go/master/LICENSE")
	}

	// Fallback 5: postman-echo.com (alternative to httpbin)
	urls = append(urls, fmt.Sprintf("https://postman-echo.com/bytes/%d", size))

	return urls
}

// testURLAccessibility tests if a URL is accessible and returns the expected content
// It returns true if the URL responds with a 2xx status code
// Uses multiple retry attempts with exponential backoff for better reliability
func testURLAccessibility(url string) bool {
	client := &http.Client{
		Timeout: 3 * time.Second, // Shorter timeout per attempt
	}

	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := client.Head(url)
		if err == nil {
			defer func() {
				if resp.Body != nil {
					_ = resp.Body.Close()
				}
			}()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return true
			}
		}

		// Wait before retry (exponential backoff)
		if attempt < maxRetries {
			time.Sleep(time.Duration(100*(attempt+1)) * time.Millisecond)
		}
	}

	return false
}

// getWorkingURL tries a list of URLs and returns the first accessible one
// If none work, returns a highly reliable fallback URL
func getWorkingURL(urls []string) string {
	// Test URLs concurrently for faster response
	type result struct {
		url   string
		works bool
	}

	results := make(chan result, len(urls))

	// Test first few URLs concurrently
	testCount := len(urls)
	if testCount > 3 {
		testCount = 3 // Limit concurrent tests for performance
	}

	for i := 0; i < testCount; i++ {
		go func(url string) {
			works := testURLAccessibility(url)
			results <- result{url: url, works: works}
		}(urls[i])
	}

	// Collect results
	for i := 0; i < testCount; i++ {
		r := <-results
		if r.works {
			return r.url
		}
	}

	// If none of the tested URLs work, try the rest sequentially
	for i := testCount; i < len(urls); i++ {
		if testURLAccessibility(urls[i]) {
			return urls[i]
		}
	}

	// Ultimate fallback: use a highly reliable GitHub URL regardless of size
	// This ensures tests don't fail due to external service unavailability
	fallback := "https://github.com/sebrandon1/grab/archive/refs/heads/main.zip"
	if testURLAccessibility(fallback) {
		return fallback
	}

	// If even GitHub is down, return the first URL to allow normal test failure
	if len(urls) > 0 {
		return urls[0]
	}

	return ""
}

// getWorkingURLs tries multiple URL lists and returns working URLs for each
func getWorkingURLs(urlLists [][]string) []string {
	var workingURLs []string
	for _, urls := range urlLists {
		workingURLs = append(workingURLs, getWorkingURL(urls))
	}
	return workingURLs
}

// Convenience functions for common test sizes
func getWorking256ByteURL() string {
	return getWorkingURL(getFallbackURLs(256))
}

func getWorking512ByteURL() string {
	return getWorkingURL(getFallbackURLs(512))
}

func getWorking1024ByteURL() string {
	return getWorkingURL(getFallbackURLs(1024))
}
