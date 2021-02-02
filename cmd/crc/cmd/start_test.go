package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/stretchr/testify/assert"
)

func TestRenderActionPlainSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&startResult{
		Success: true,
		ClusterConfig: &clusterConfig{
			URL: constants.DefaultAPIURL,
			AdminCredentials: credentials{
				Username: "kubeadmin",
				Password: "secret",
			},
			DeveloperCredentials: credentials{
				Username: "developer",
				Password: "developer",
			},
		},
	}, out, ""))
	assert.Equal(t, `Started the OpenShift cluster

To access the cluster, first set up your environment by following the instructions returned by executing 'crc oc-env'.
Then you can access your cluster by running 'oc login -u developer -p developer https://api.crc.testing:6443'.
To login as a cluster admin, run 'oc login -u kubeadmin -p secret https://api.crc.testing:6443'.

You can also run 'crc console' and use the above credentials to access the OpenShift web console.
The console will open in your default browser.
`, out.String())
}

func TestRenderActionPlainFailure(t *testing.T) {
	out := new(bytes.Buffer)
	err := errors.New("broken")
	assert.EqualError(t, render(&startResult{
		Success: false,
		Error:   crcErrors.ToSerializableError(err),
	}, out, ""), "broken")
	assert.Equal(t, "", out.String())
}

func TestRenderActionJSONSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&startResult{
		Success: true,
		ClusterConfig: &clusterConfig{
			WebConsoleURL: constants.DefaultWebConsoleURL,
			URL:           constants.DefaultAPIURL,
			AdminCredentials: credentials{
				Username: "kubeadmin",
				Password: "secret",
			},
			DeveloperCredentials: credentials{
				Username: "developer",
				Password: "developer",
			},
		},
	}, out, jsonFormat))
	assert.Equal(t, `{
  "success": true,
  "clusterConfig": {
    "cacert": "",
    "webConsoleUrl": "https://console-openshift-console.apps-crc.testing",
    "url": "https://api.crc.testing:6443",
    "adminCredentials": {
      "username": "kubeadmin",
      "password": "secret"
    },
    "developerCredentials": {
      "username": "developer",
      "password": "developer"
    }
  }
}
`, out.String())
}

func TestRenderActionJSONFailure(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&startResult{
		Success: false,
		Error:   crcErrors.ToSerializableError(errors.New("broken")),
	}, out, jsonFormat))
	assert.JSONEq(t, `{"success": false, "error": "broken"}`, out.String())
}
