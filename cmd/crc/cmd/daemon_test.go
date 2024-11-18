package cmd

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogResponseBodyLogsResponseBodyForFailedResponseCodes(t *testing.T) {
	// Given
	var logBuffer bytes.Buffer
	var responseBuffer bytes.Buffer
	responseBuffer.WriteString("{\"status\": \"FAILURE\"}")
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(os.Stdout)
	requestURL, err := url.Parse("http://127.0.0.1/log")
	assert.NoError(t, err)
	httpRequest := &http.Request{
		Method: "GET",
		URL:    requestURL,
	}

	// When
	logResponseBodyConditionally(500, &responseBuffer, httpRequest)

	// Then
	assert.Greater(t, logBuffer.Len(), 0)
	assert.Contains(t, logBuffer.String(), ("\\\"GET /log\\\" Response Body: {\\\"status\\\": \\\"FAILURE\\\"}"))
}

func TestLogResponseBodyLogsNothingWhenResponseSuccessful(t *testing.T) {
	// Given
	var logBuffer bytes.Buffer
	var responseBuffer bytes.Buffer
	responseBuffer.WriteString("{\"status\": \"SUCCESS\"}")
	logrus.SetOutput(&logBuffer)
	defer logrus.SetOutput(os.Stdout)
	requestURL, err := url.Parse("http://127.0.0.1/log")
	assert.NoError(t, err)
	httpRequest := &http.Request{
		Method: "GET",
		URL:    requestURL,
	}

	// When
	logResponseBodyConditionally(200, &responseBuffer, httpRequest)

	// Then
	assert.Equal(t, logBuffer.Len(), 0)
}
