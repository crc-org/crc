package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/stretchr/testify/assert"
)

func TestConsolePlainSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, fakemachine.NewClient(), true, false, ""))
	assert.Equal(t, fmt.Sprintf("%s\n", fakemachine.DummyClusterConfig.WebConsoleURL), out.String())
}

func TestConsolePlainError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.EqualError(t, runConsole(out, fakemachine.NewFailingClient(), true, false, ""), "console failed")
}

func TestConsoleWithPrintCredentialsPlainSuccess(t *testing.T) {
	expectedOut := fmt.Sprintf(`To login as a regular user, run 'oc login -u developer -p developer %s'.
To login as an admin, run 'oc login -u kubeadmin -p %s %s'
`, fakemachine.DummyClusterConfig.ClusterAPI, fakemachine.DummyClusterConfig.KubeAdminPass, fakemachine.DummyClusterConfig.ClusterAPI)
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, fakemachine.NewClient(), false, true, ""))
	assert.Equal(t, expectedOut, out.String())
}

func TestConsoleJSONSuccess(t *testing.T) {
	expectedJSONOut := fmt.Sprintf(`{
  "success": true,
  "clusterConfig": {
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
}`, fakemachine.DummyClusterConfig.WebConsoleURL, fakemachine.DummyClusterConfig.ClusterAPI, fakemachine.DummyClusterConfig.KubeAdminPass)
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, fakemachine.NewClient(), false, false, jsonFormat))
	assert.JSONEq(t, expectedJSONOut, out.String())
}

func TestConsoleJSONError(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, runConsole(out, fakemachine.NewFailingClient(), false, false, jsonFormat))
	assert.JSONEq(t, `{"error":"console failed", "success":false}`, out.String())
}
