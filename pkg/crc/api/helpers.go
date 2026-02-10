package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/crc-org/crc/v2/pkg/crc/api/client"
	"github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

type context struct {
	method      string
	requestBody []byte
	url         *url.URL

	code         int
	headers      map[string]string
	responseBody []byte
}

func (c *context) Bind(r interface{}) error {
	return json.Unmarshal(c.requestBody, r)
}

func (c *context) JSON(code int, r interface{}) error {
	c.code = code
	var err error
	c.responseBody, err = json.Marshal(r)
	c.headers["Content-Type"] = "application/json"
	return err
}

func (c *context) String(code int, r string) error {
	c.code = code
	c.responseBody = []byte(r)
	c.headers["Content-Type"] = "text/plain; charset=UTF-8"
	return nil
}

func (c *context) Code(code int) error {
	c.code = code
	return nil
}

type server struct {
	routes     map[string]map[string]func(*context) error
	routesLock sync.RWMutex
}

func newServer() *server {
	return &server{
		routes: make(map[string]map[string]func(*context) error),
	}
}

func (s *server) GET(pattern string, handler func(c *context) error) {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	if _, ok := s.routes[pattern]; !ok {
		s.routes[pattern] = make(map[string]func(*context) error)
	}
	s.routes[pattern][http.MethodGet] = handler
}

func (s *server) POST(pattern string, handler func(c *context) error) {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	if _, ok := s.routes[pattern]; !ok {
		s.routes[pattern] = make(map[string]func(*context) error)
	}
	s.routes[pattern][http.MethodPost] = handler
}

func (s *server) DELETE(pattern string, handler func(c *context) error) {
	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	if _, ok := s.routes[pattern]; !ok {
		s.routes[pattern] = make(map[string]func(*context) error)
	}
	s.routes[pattern][http.MethodDelete] = handler
}

func (s *server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.routesLock.RLock()
		route, ok := s.routes[r.URL.Path]
		if !ok {
			s.routesLock.RUnlock()
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		handler, ok := route[r.Method]
		if !ok {
			s.routesLock.RUnlock()
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		s.routesLock.RUnlock()

		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c := &context{
			method:      r.Method,
			requestBody: requestBody,
			headers:     make(map[string]string),
			url:         r.URL,
		}
		if err := handler(c); err != nil {
			if multiErr, ok := err.(errors.MultiError); ok {
				writeMultiError(w, multiErr)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(c.code)
		for k, v := range c.headers {
			w.Header().Set(k, v)
		}
		if _, err := w.Write(c.responseBody); err != nil {
			logging.Error("Failed to send response: ", err)
		}
	})
}

func writeMultiError(w http.ResponseWriter, multiErr errors.MultiError) {
	validationErrors := make([]client.ValidationError, 0, len(multiErr.Errors))
	for _, e := range multiErr.Errors {
		validationErrors = append(validationErrors, client.ValidationError{
			Message: e.Error(),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(validationErrors); err != nil {
		logging.Error("Failed to write error response: ", err)
	}
}
