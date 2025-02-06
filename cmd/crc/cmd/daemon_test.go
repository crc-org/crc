package cmd

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/api/client"

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

func TestCheckDaemonVersion_WhenNoErrorWhileFetchingVersion_ThenThrowDaemonAlreadyStartedError(t *testing.T) {
	// Given
	daemonVersionSupplier = func() (client.VersionResult, error) {
		return client.VersionResult{}, nil
	}

	// When
	result, err := checkDaemonVersion()

	// Then
	assert.Equal(t, true, result)
	assert.Errorf(t, err, "daemon has been started in the background")
}

func TestCheckDaemonVersion_WhenErrorReturnedWhileFetchingVersion_ThenReturnFalse(t *testing.T) {
	// Given
	daemonVersionSupplier = func() (client.VersionResult, error) {
		return client.VersionResult{}, errors.New("daemon not started")
	}

	// When
	result, err := checkDaemonVersion()

	// Then
	assert.NoError(t, err)
	assert.Equal(t, false, result)
}
