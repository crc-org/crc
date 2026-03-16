package cmd

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/crc-org/crc/v2/pkg/crc/api/client"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"

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

type fakeHostsFileEditor struct {
	addCalled    bool
	removeCalled bool
	addIP        string
	addHosts     []string
	removeHosts  []string
}

func (fake *fakeHostsFileEditor) Add(ip string, hostnames ...string) error {
	fake.addCalled = true
	fake.addIP = ip
	fake.addHosts = append([]string(nil), hostnames...)
	return nil
}

func (fake *fakeHostsFileEditor) Remove(hostnames ...string) error {
	fake.removeCalled = true
	fake.removeHosts = append([]string(nil), hostnames...)
	return nil
}

func TestGatewayAPIMux_HostsEndpointsRespectModifyHostsFile(t *testing.T) {
	tests := []struct {
		name             string
		modifyHostsFile  bool
		path             string
		expectAddCalled  bool
		expectRemoveCall bool
	}{
		{
			name:            "add-enabled",
			modifyHostsFile: true,
			path:            "/hosts/add",
			expectAddCalled: true,
		},
		{
			name:            "add-disabled",
			modifyHostsFile: false,
			path:            "/hosts/add",
		},
		{
			name:             "remove-enabled",
			modifyHostsFile:  true,
			path:             "/hosts/remove",
			expectRemoveCall: true,
		},
		{
			name:            "remove-disabled",
			modifyHostsFile: false,
			path:            "/hosts/remove",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			cfg := crcConfig.New(crcConfig.NewEmptyInMemoryStorage(), crcConfig.NewEmptyInMemorySecretStorage())
			crcConfig.RegisterSettings(cfg)
			_, err := cfg.Set(crcConfig.ModifyHostsFile, test.modifyHostsFile)
			assert.NoError(t, err)
			hostsEditor := &fakeHostsFileEditor{}
			mux := gatewayAPIMux(cfg, hostsEditor)
			rec := httptest.NewRecorder()

			// When
			req := httptest.NewRequest(http.MethodPost, test.path, bytes.NewBufferString(`["api.crc.testing"]`))
			mux.ServeHTTP(rec, req)

			// Then
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, test.expectAddCalled, hostsEditor.addCalled)
			assert.Equal(t, test.expectRemoveCall, hostsEditor.removeCalled)
		})
	}
}
