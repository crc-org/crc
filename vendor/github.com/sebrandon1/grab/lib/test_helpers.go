package lib

import (
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

// getFallbackURLs returns a list of fallback URLs for a given size in bytes.
//
// Small, static files on raw.githubusercontent.com come first: they are cheap
// and are not rate-limited for shared CI egress. httpbin.org and
// postman-echo.com were previously used here but answer HEAD with 200 and then
// return 429 Too Many Requests to GitHub Actions runners, which made the
// real-HTTP tests flaky. They are intentionally omitted.
func getFallbackURLs(size int) []string {
	urls := []string{}

	// size is nominal: every candidate below is non-empty and downloads
	// quickly, so tests only need a reliably reachable body.
	switch size {
	case 256:
		urls = append(urls,
			"https://raw.githubusercontent.com/actions/checkout/main/.gitignore",
			"https://raw.githubusercontent.com/spf13/cobra/main/go.mod",
			"https://raw.githubusercontent.com/github/gitignore/main/Go.gitignore")
	case 512:
		urls = append(urls,
			"https://raw.githubusercontent.com/github/gitignore/main/Go.gitignore",
			"https://raw.githubusercontent.com/golang/go/master/SECURITY.md",
			"https://raw.githubusercontent.com/golang/go/master/LICENSE")
	case 768:
		urls = append(urls,
			"https://raw.githubusercontent.com/golang/go/master/LICENSE",
			"https://raw.githubusercontent.com/golang/go/master/README.md",
			"https://raw.githubusercontent.com/kubernetes/kubernetes/master/.gitignore")
	case 1024:
		urls = append(urls,
			"https://raw.githubusercontent.com/golang/go/master/README.md",
			"https://raw.githubusercontent.com/golang/go/master/src/go/doc/comment.go",
			"https://raw.githubusercontent.com/kubernetes/kubernetes/master/README.md")
	case 2048:
		urls = append(urls,
			"https://raw.githubusercontent.com/golang/go/master/src/go/doc/comment.go",
			"https://raw.githubusercontent.com/kubernetes/kubernetes/master/README.md",
			"https://raw.githubusercontent.com/golang/go/master/src/cmd/go/testdata/script/mod_tidy_compat.txt")
	default:
		urls = append(urls,
			"https://raw.githubusercontent.com/golang/go/master/LICENSE",
			"https://raw.githubusercontent.com/golang/go/master/README.md",
			"https://raw.githubusercontent.com/microsoft/vscode/main/package.json")
	}

	// Last-resort fallbacks on other hosts: small, stable, and cheap to
	// download even when the nominal test file is only a few hundred bytes.
	urls = append(urls,
		"https://go.dev/VERSION?m=text",
		"https://proxy.golang.org/golang.org/x/text/@v/list")

	return urls
}

// testHTTPClient is shared so the probes started by getWorkingURL reuse
// connections instead of opening a new transport per URL.
var testHTTPClient = &http.Client{Timeout: 3 * time.Second}

// testURLAccessibility reports whether url currently serves a 2xx response to
// the same GET request the download code will make. Probing with GET (rather
// than HEAD) matters: some hosts answer HEAD successfully but rate-limit GET.
// The body is closed immediately, so only the response headers are consumed.
func testURLAccessibility(url string) bool {
	const maxAttempts = 2
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := testHTTPClient.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return true
			}
		}

		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}

	return false
}

// workingURLCache memoizes the working URL chosen for each fallback list so a
// test run probes each host at most once. resetWorkingURLs clears it so a
// retrying test can recover if a cached host starts rate-limiting.
var workingURLCache sync.Map // map[string]string

// getWorkingURL returns the first URL in list order that is currently
// accessible. Probes run concurrently for speed, but the choice is the
// lowest-index success, so it is deterministic regardless of which host
// responds first. Successful choices are cached for the rest of the test run.
func getWorkingURL(urls []string) string {
	if len(urls) == 0 {
		return ""
	}

	cacheKey := strings.Join(urls, "\n")
	if cached, ok := workingURLCache.Load(cacheKey); ok {
		return cached.(string)
	}

	type probe struct {
		index int
		works bool
	}
	probes := make(chan probe, len(urls))
	for i, url := range urls {
		go func(i int, url string) {
			probes <- probe{index: i, works: testURLAccessibility(url)}
		}(i, url)
	}

	best := -1
	for range urls {
		if p := <-probes; p.works && (best == -1 || p.index < best) {
			best = p.index
		}
	}

	if best >= 0 {
		workingURLCache.Store(cacheKey, urls[best])
		return urls[best]
	}

	// Ultimate fallback on a separate host, independent of the list above.
	// This ensures tests don't fail due to external service unavailability.
	const ultimateFallback = "https://go.dev/VERSION?m=text"
	if testURLAccessibility(ultimateFallback) {
		return ultimateFallback
	}

	// If nothing is reachable, return the first URL to allow normal test
	// failure. Do not cache this: a later call may succeed.
	return urls[0]
}

// resetWorkingURLs drops the memoized URL choices so the next getWorkingURL
// call re-probes the hosts. Tests call it before retrying a failed attempt.
func resetWorkingURLs() {
	workingURLCache.Range(func(key, _ any) bool {
		workingURLCache.Delete(key)
		return true
	})
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
