package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type context struct {
	method      string
	requestBody []byte

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
	var err error
	c.responseBody = []byte(r)
	c.headers["Content-Type"] = "text/plain; charset=UTF-8"
	return err
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

		requestBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		c := &context{
			method:      r.Method,
			requestBody: requestBody,
			headers:     make(map[string]string),
		}
		if err := handler(c); err != nil {
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
