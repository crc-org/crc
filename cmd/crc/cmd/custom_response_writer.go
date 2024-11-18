package cmd

import (
	"bytes"
	"net/http"
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
