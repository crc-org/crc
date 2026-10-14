package apilog

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"net"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type contextKey string

const fieldsContextKey = contextKey("apilog_fields")

// maxCapturedErrorBody caps how many bytes of an error response body
const maxCapturedErrorBody = 4096

// AddField adds a key-value pair to the audit log for this request.
// If the request was not wrapped by Middleware, this is a no-op.
func AddField(r *http.Request, key string, value interface{}) {
	if fields, ok := r.Context().Value(fieldsContextKey).(log.Fields); ok {
		fields[key] = value
	}
}

// SetError records an error message in the audit log for this request.
func SetError(r *http.Request, err error) {
	if err == nil {
		return
	}
	if fields, ok := r.Context().Value(fieldsContextKey).(log.Fields); ok {
		fields["error"] = err.Error()
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	errBody    strings.Builder
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	if rw.statusCode >= http.StatusBadRequest && rw.errBody.Len() < maxCapturedErrorBody {
		remaining := maxCapturedErrorBody - rw.errBody.Len()
		remaining = min(remaining, len(b))
		rw.errBody.Write(b[:remaining])
	}
	return rw.ResponseWriter.Write(b)
}

// Hijack is needed for endpoints like /tunnel and /connect.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("webserver doesn't support hijacking")
	}
	return hijacker.Hijack()
}

// Middleware is an HTTP middleware that logs API requests with any extra
// fields added by handlers via AddField or SetError. If a request's status
// code indicates an error and no handler called SetError explicitly,
// Middleware falls back to using the response body (e.g. the message passed
// to http.Error) as the error field.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extra := make(log.Fields)
		ctx := context.WithValue(r.Context(), fieldsContextKey, extra)
		r = r.WithContext(ctx)

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
		}

		next.ServeHTTP(rw, r)

		outcome := "success"
		if rw.statusCode >= http.StatusBadRequest {
			outcome = "error"
			if _, hasErr := extra["error"]; !hasErr {
				if msg := strings.TrimSpace(rw.errBody.String()); msg != "" {
					SetError(r, fmt.Errorf("%s", msg))
				}
			}
		}
		if _, hasErr := extra["error"]; hasErr {
			outcome = "error"
		}

		fields := make(log.Fields)
		maps.Copy(fields, extra)
		fields["component"] = "services-api"
		fields["endpoint"] = r.URL.Path
		fields["method"] = r.Method
		fields["source"] = r.RemoteAddr
		fields["outcome"] = outcome

		log.WithFields(fields).Info("gvproxy API request")
	})
}
