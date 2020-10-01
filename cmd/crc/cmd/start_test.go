package cmd

import (
	"bytes"
	"testing"

	"github.com/code-ready/crc/pkg/crc/constants"
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

To access the cluster, first set up your environment by following 'crc oc-env' instructions.
Then you can access it by running 'oc login -u developer -p developer https://api.crc.testing:6443'.
To login as an admin, run 'oc login -u kubeadmin -p secret https://api.crc.testing:6443'.
To access the cluster, first set up your environment by following 'crc oc-env' instructions.

You can now run 'crc console' and use these credentials to access the OpenShift web console.
`, out.String())
}

func TestRenderActionPlainFailure(t *testing.T) {
	out := new(bytes.Buffer)
	assert.EqualError(t, render(&startResult{
		Success: false,
		Error:   "broken",
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
		Error:   "broken",
	}, out, jsonFormat))
	assert.JSONEq(t, `{"success": false, "error": "broken"}`, out.String())
}
