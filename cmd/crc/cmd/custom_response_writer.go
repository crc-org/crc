package cmd

import (
	"bytes"
	"io"
	"net/http"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

// CustomResponseWriter wraps the standard http.ResponseWriter and captures the response body
type CustomResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

// NewCustomResponseWriter creates a new CustomResponseWriter
func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

// WriteHeader allows capturing and modifying the status code
func (rw *CustomResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body and logs it
func (rw *CustomResponseWriter) Write(p []byte) (int, error) {
	bufferLen, err := rw.body.Write(p)
	if err != nil {
		return bufferLen, err
	}

	return rw.ResponseWriter.Write(p)
}

// interceptResponseBodyMiddleware injects the custom bodyConsumer function (received as second argument) into
// http.HandleFunc logic that allows users to intercept response body as per their requirements (e.g. logging)
// and returns updated http.Handler
func interceptResponseBodyMiddleware(next http.Handler, bodyConsumer func(statusCode int, buffer *bytes.Buffer, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseWriter := NewCustomResponseWriter(w)
		next.ServeHTTP(responseWriter, r)
		bodyConsumer(responseWriter.statusCode, responseWriter.body, r)
	})
}

// logRequestMiddleware logs every request to the daemon log file.
func logRequestMiddleware(next http.Handler, label string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := readRequestBodyForLogging(r)
		if err != nil {
			logging.Warnf("failed to read request body for %s %s: %v", r.Method, r.URL.Path, err)
		}

		responseWriter := NewCustomResponseWriter(w)
		next.ServeHTTP(responseWriter, r)
		logRequest(label, responseWriter.statusCode, r, requestBody)
	})
}

func readRequestBodyForLogging(r *http.Request) ([]byte, error) {
	if r.Method != http.MethodPost || r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func logRequest(label string, statusCode int, r *http.Request, requestBody []byte) {
	fields := map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"status": statusCode,
	}
	if r.Method == http.MethodPost {
		fields["body"] = string(requestBody)
	}
	entry := logging.WithFields(fields)
	if statusCode >= http.StatusBadRequest {
		entry.Warn(label)
		return
	}
	entry.Info(label)
}
