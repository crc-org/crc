package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type logrusCaptureHook struct {
	entries []logrus.Entry
}

func (h *logrusCaptureHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *logrusCaptureHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, *entry)
	return nil
}

func withLogrusCaptureHook(t *testing.T) *logrusCaptureHook {
	t.Helper()
	hook := &logrusCaptureHook{}
	logrus.AddHook(hook)
	t.Cleanup(func() {
		logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))
	})
	return hook
}

type TestHandler struct {
}

func (t *TestHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprint(w, "Testing!")
	if err != nil {
		return
	}
}

func TestLogResponseBodyMiddlewareCapturesResponseAsExpected(t *testing.T) {
	// Given
	interceptedResponseStatusCode := -1
	interceptedResponseBody := ""
	responseBodyConsumer := func(statusCode int, buffer *bytes.Buffer, _ *http.Request) {
		interceptedResponseStatusCode = statusCode
		interceptedResponseBody = buffer.String()
	}
	testHandler := &TestHandler{}
	server := httptest.NewServer(interceptResponseBodyMiddleware(http.StripPrefix("/", testHandler), responseBodyConsumer))
	defer server.Close()
	// When
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	// Then
	responseBody := new(bytes.Buffer)
	bytesRead, err := responseBody.ReadFrom(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, 200, interceptedResponseStatusCode)
	assert.Equal(t, int64(8), bytesRead)
	assert.Equal(t, "Testing!", responseBody.String())
	assert.Equal(t, "Testing!", interceptedResponseBody)
}

func TestLogRequestMiddlewareLogsSuccessfulRequests(t *testing.T) {
	var logBuffer bytes.Buffer
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(os.Stdout)

	handler := logRequestMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), "network request")

	req := httptest.NewRequest(http.MethodGet, "/services/forwarder/all", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, logBuffer.String(), "network request")
	assert.Contains(t, logBuffer.String(), "/services/forwarder/all")
}

func TestLogRequestMiddlewareLogsPostRequestBody(t *testing.T) {
	hook := withLogrusCaptureHook(t)

	var receivedBody string
	handler := logRequestMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}), "network request")

	reqBody := `{"local":"127.0.0.1:2224","remote":"192.168.127.2:22","protocol":"tcp"}`
	req := httptest.NewRequest(http.MethodPost, "/services/forwarder/expose", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, reqBody, receivedBody)
	assert.Len(t, hook.entries, 1)
	entry := hook.entries[0]
	assert.Equal(t, "network request", entry.Message)
	assert.Equal(t, http.MethodPost, entry.Data["method"])
	assert.Equal(t, "/services/forwarder/expose", entry.Data["path"])
	assert.Equal(t, http.StatusOK, entry.Data["status"])
	assert.Equal(t, reqBody, entry.Data["body"])
}

func TestLogRequestMiddlewareLogsFailedRequestsAsWarning(t *testing.T) {
	var logBuffer bytes.Buffer
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(os.Stdout)

	handler := logRequestMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}), "network request")

	req := httptest.NewRequest(http.MethodPost, "/services/forwarder/expose", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, logBuffer.String(), "level=warning")
	assert.Contains(t, logBuffer.String(), "network request")
}
