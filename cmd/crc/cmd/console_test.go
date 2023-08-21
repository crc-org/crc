package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	apiTypes "github.com/crc-org/crc/v2/pkg/crc/api/client"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	"github.com/crc-org/crc/v2/pkg/crc/machine/fakemachine"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	mocks "github.com/crc-org/crc/v2/test/mocks/api"
	"github.com/stretchr/testify/assert"
)

var DummyClusterConfig = types.ClusterConfig{
	ClusterType:   "openshift",
	ClusterCACert: "MIIDODCCAiCgAwIBAgIIRVfCKNUa1wIwDQYJ",
	KubeConfig:    "/tmp/kubeconfig",
	KubeAdminPass: "foobar",
	ClusterAPI:    "https://foo.testing:6443",
	WebConsoleURL: "https://console.foo.testing:6443",
	ProxyConfig:   nil,
}

func setUpClientForConsole(t *testing.T) *daemonclient.Client {
	client := mocks.NewClient(t)

	client.On("WebconsoleURL").Return(
		&apiTypes.ConsoleResult{
			ClusterConfig: DummyClusterConfig,
			State:         state.Running,
		}, nil)
	return &daemonclient.Client{
		APIClient: client,
	}
}

func setUpFailingClientForConsole(t *testing.T) *daemonclient.Client {
	client := mocks.NewClient(t)

	client.On("WebconsoleURL").Return(
		nil, errors.New("console failed"))
	return &daemonclient.Client{
		APIClient: client,
	}
}

func TestConsolePlainSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, setUpClientForConsole(t), true, false, ""))
	assert.Equal(t, fmt.Sprintf("%s\n", fakemachine.DummyClusterConfig.WebConsoleURL), out.String())
}

func TestConsolePlainError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.EqualError(t, runConsole(out, setUpFailingClientForConsole(t), true, false, ""), "console failed")
}

func TestConsoleWithPrintCredentialsPlainSuccess(t *testing.T) {
	expectedOut := fmt.Sprintf(`To login as a regular user, run 'oc login -u developer -p developer %s'.
To login as an admin, run 'oc login -u kubeadmin -p %s %s'
`, fakemachine.DummyClusterConfig.ClusterAPI, fakemachine.DummyClusterConfig.KubeAdminPass, fakemachine.DummyClusterConfig.ClusterAPI)
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, setUpClientForConsole(t), false, true, ""))
	assert.Equal(t, expectedOut, out.String())
}

func TestConsoleWithPrintCredentialsAndURLPlainSuccess(t *testing.T) {
	expectedOut := fmt.Sprintf(`%s
To login as a regular user, run 'oc login -u developer -p developer %s'.
To login as an admin, run 'oc login -u kubeadmin -p %s %s'
`, fakemachine.DummyClusterConfig.WebConsoleURL, fakemachine.DummyClusterConfig.ClusterAPI, fakemachine.DummyClusterConfig.KubeAdminPass, fakemachine.DummyClusterConfig.ClusterAPI)
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, setUpClientForConsole(t), true, true, ""))
	assert.Equal(t, expectedOut, out.String())
}

func TestConsoleJSONSuccess(t *testing.T) {
	expectedJSONOut := fmt.Sprintf(`{
  "success": true,
  "clusterConfig": {
    "clusterType": "openshift",
    "cacert": "%s",
    "webConsoleUrl": "%s",
    "url": "%s",
    "adminCredentials": {
      "username": "kubeadmin",
      "password": "%s"
    },
    "developerCredentials": {
      "username": "developer",
      "password": "developer"
    }
  }
}`, fakemachine.DummyClusterConfig.ClusterCACert, fakemachine.DummyClusterConfig.WebConsoleURL, fakemachine.DummyClusterConfig.ClusterAPI, fakemachine.DummyClusterConfig.KubeAdminPass)
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, setUpClientForConsole(t), false, false, jsonFormat))
	assert.JSONEq(t, expectedJSONOut, out.String())
}

func TestConsoleJSONError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, setUpFailingClientForConsole(t), false, false, jsonFormat))
	assert.JSONEq(t, `{"error":"console failed", "success":false}`, out.String())
}
