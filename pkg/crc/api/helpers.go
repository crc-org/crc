package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type apiVersion string

const (
	apiV1 = apiVersion("v1") // the default version when no version is provided
	apiV2 = apiVersion("v2")

	// AcceptVersionHeaderKey is the header key of "Accept-Version".
	AcceptVersionHeaderKey = "Accept-Version"
	// AcceptHeaderKey is the header key of "Accept".
	AcceptHeaderKey = "Accept"
	// AcceptHeaderVersionValue is the Accept's header value search term the requested version.
	AcceptHeaderVersionValue = "version"

	NotFound = apiVersion("notfound")
)

func (v apiVersion) isValid() bool {
	return v == apiV1 || v == apiV2
}

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
	routes     map[string]map[string]map[apiVersion]func(*context) error
	routesLock sync.RWMutex
}

func newServer() *server {
	return &server{
		routes: make(map[string]map[string]map[apiVersion]func(*context) error),
	}
}

func (s *server) GET(pattern string, handler func(c *context) error, version apiVersion) {
	if !version.isValid() {
		panic("Invalid version used to register handler")
	}
	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	if _, ok := s.routes[pattern]; !ok {
		s.routes[pattern] = make(map[string]map[apiVersion]func(*context) error)
	}
	if _, ok := s.routes[pattern][http.MethodGet]; !ok {
		s.routes[pattern][http.MethodGet] = make(map[apiVersion]func(*context) error)
	}
	s.routes[pattern][http.MethodGet][version] = handler
}

func (s *server) POST(pattern string, handler func(c *context) error, version apiVersion) {
	if !version.isValid() {
		panic("Invalid version used to register handler")
	}
	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	if _, ok := s.routes[pattern]; !ok {
		s.routes[pattern] = make(map[string]map[apiVersion]func(*context) error)
	}
	if _, ok := s.routes[pattern][http.MethodPost]; !ok {
		s.routes[pattern][http.MethodPost] = make(map[apiVersion]func(*context) error)
	}
	s.routes[pattern][http.MethodPost][version] = handler
}

func (s *server) DELETE(pattern string, handler func(c *context) error, version apiVersion) {
	if !version.isValid() {
		panic("Invalid version used to register handler")
	}
	s.routesLock.Lock()
	defer s.routesLock.Unlock()
	if _, ok := s.routes[pattern]; !ok {
		s.routes[pattern] = make(map[string]map[apiVersion]func(*context) error)
	}
	if _, ok := s.routes[pattern][http.MethodDelete]; !ok {
		s.routes[pattern][http.MethodDelete] = make(map[apiVersion]func(*context) error)
	}
	s.routes[pattern][http.MethodDelete][version] = handler
}

func (s *server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.routesLock.RLock()
		var version = apiV1 // sets v1 as the default in case no version in req
		if v := getVersion(r); v != NotFound {
			version = v
		}
		route, ok := s.routes[r.URL.Path]
		if !ok {
			s.routesLock.RUnlock()
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		versionedHandlers, ok := route[r.Method]
		if !ok {
			s.routesLock.RUnlock()
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		handler, ok := versionedHandlers[version]
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
			url:         r.URL,
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

// GetVersion returns the current request version.
//
// By default `GetVersion` will try to read from:
// - "Accept" header, i.e Accept: "application/json; version=1.0"
// - "Accept-Version" header, i.e Accept-Version: "1.0"
func getVersion(r *http.Request) apiVersion {
	// firstly by the "Accept-Version" header.
	if version := r.Header.Get(AcceptVersionHeaderKey); version != "" {
		return apiVersion(version)
	}

	// secondly by the "Accept" header which is like "...; version=1.0"
	acceptValue := r.Header.Get(AcceptHeaderKey)
	if acceptValue != "" {
		if idx := strings.Index(acceptValue, AcceptHeaderVersionValue); idx != -1 {
			rem := acceptValue[idx:]
			startVersion := strings.Index(rem, "=")
			if startVersion == -1 || len(rem) < startVersion+1 {
				return NotFound
			}

			rem = rem[startVersion+1:]

			end := strings.Index(rem, " ")
			if end == -1 {
				end = strings.Index(rem, ";")
			}
			if end == -1 {
				end = len(rem)
			}

			if version := rem[:end]; version != "" {
				return apiVersion(version)
			}
		}
	}

	return NotFound
}
