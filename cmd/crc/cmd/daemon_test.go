package cmd

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
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

func TestCreateNewVirtualNetworkConfig(t *testing.T) {
	// Given
	oldPcapFileEnvVal := os.Getenv("CRC_DAEMON_PCAP_FILE")
	err := os.Setenv("CRC_DAEMON_PCAP_FILE", "/tmp/pcapfile")
	assert.NoError(t, err)
	defer func(key, value string) {
		err := os.Setenv(key, value)
		assert.NoError(t, err)
	}("CRC_DAEMON_PCAP_FILE", oldPcapFileEnvVal)
	testCrcConfig := crcConfig.New(crcConfig.NewEmptyInMemoryStorage(), crcConfig.NewEmptyInMemorySecretStorage())

	// When
	virtualNetworkConfig := createNewVirtualNetworkConfig(testCrcConfig)

	// Then
	assert.Equal(t, false, virtualNetworkConfig.Debug)
	assert.Equal(t, "/tmp/pcapfile", virtualNetworkConfig.CaptureFile)
	assert.Equal(t, 4000, virtualNetworkConfig.MTU)
	assert.Equal(t, "192.168.127.0/24", virtualNetworkConfig.Subnet)
	assert.Equal(t, "192.168.127.1", virtualNetworkConfig.GatewayIP)
	assert.ElementsMatch(t, []string{"192.168.127.254"}, virtualNetworkConfig.GatewayVirtualIPs)
	assert.Equal(t, "5a:94:ef:e4:0c:dd", virtualNetworkConfig.GatewayMacAddress)
	assert.Equal(t, types.Protocol("hyperkit"), virtualNetworkConfig.Protocol)

	assert.Len(t, virtualNetworkConfig.DHCPStaticLeases, 1)
	assert.Equal(t, "5a:94:ef:e4:0c:ee", virtualNetworkConfig.DHCPStaticLeases["192.168.127.2"])

	assert.Len(t, virtualNetworkConfig.DNS, 4)
	assert.Equal(t, "apps-crc.testing.", virtualNetworkConfig.DNS[0].Name)
	assert.Equal(t, net.ParseIP("192.168.127.2"), virtualNetworkConfig.DNS[0].DefaultIP)
	assert.Equal(t, "crc.testing.", virtualNetworkConfig.DNS[1].Name)
	assert.Equal(t, "host", virtualNetworkConfig.DNS[1].Records[0].Name)
	assert.Equal(t, net.ParseIP("192.168.127.254"), virtualNetworkConfig.DNS[1].Records[0].IP)
	assert.Equal(t, "gateway", virtualNetworkConfig.DNS[1].Records[1].Name)
	assert.Equal(t, net.ParseIP("192.168.127.1"), virtualNetworkConfig.DNS[1].Records[1].IP)
	assert.Equal(t, "api", virtualNetworkConfig.DNS[1].Records[2].Name)
	assert.Equal(t, net.ParseIP("192.168.127.2"), virtualNetworkConfig.DNS[1].Records[2].IP)
	assert.Equal(t, "api-int", virtualNetworkConfig.DNS[1].Records[3].Name)
	assert.Equal(t, net.ParseIP("192.168.127.2"), virtualNetworkConfig.DNS[1].Records[3].IP)
	assert.Equal(t, regexp.MustCompile("crc-(.*?)-master-0"), virtualNetworkConfig.DNS[1].Records[4].Regexp)
	assert.Equal(t, net.ParseIP("192.168.126.11"), virtualNetworkConfig.DNS[1].Records[4].IP)

	assert.Equal(t, "containers.internal.", virtualNetworkConfig.DNS[2].Name)
	assert.Len(t, virtualNetworkConfig.DNS[2].Records, 1)
	assert.Equal(t, "gateway", virtualNetworkConfig.DNS[2].Records[0].Name)
	assert.Equal(t, net.ParseIP("192.168.127.254"), virtualNetworkConfig.DNS[2].Records[0].IP)
	assert.Equal(t, "docker.internal.", virtualNetworkConfig.DNS[3].Name)
	assert.Len(t, virtualNetworkConfig.DNS[3].Records, 1)
	assert.Equal(t, "gateway", virtualNetworkConfig.DNS[3].Records[0].Name)
	assert.Equal(t, net.ParseIP("192.168.127.254"), virtualNetworkConfig.DNS[3].Records[0].IP)
}

func TestCreateNewVirtualNetworkConfig_WhenHostNetworkConfigSet_ThenSetNAT(t *testing.T) {
	// Given
	testCrcConfig := crcConfig.New(crcConfig.NewEmptyInMemoryStorage(), crcConfig.NewEmptyInMemorySecretStorage())
	testCrcConfig.AddSetting("host-network-access", false, crcConfig.ValidateBool, crcConfig.SuccessfullyApplied, "test message")
	_, err := testCrcConfig.Set(crcConfig.HostNetworkAccess, true)
	assert.NoError(t, err)

	// When
	virtualNetworkConfig := createNewVirtualNetworkConfig(testCrcConfig)

	// Then
	assert.Equal(t, "127.0.0.1", virtualNetworkConfig.NAT["192.168.127.254"])
}
