package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
