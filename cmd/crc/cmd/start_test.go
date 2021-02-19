package cmd

import (
	"bytes"
	"errors"
	"runtime"
	"testing"

	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/os/shell"
	"github.com/stretchr/testify/assert"
)

const (
	defaultWebConsoleURL = "https://console-openshift-console.apps-crc.testing"
	defaultAPIURL        = "https://api.crc.testing:6443"
)

func TestRenderActionPlainSuccess(t *testing.T) {
	out := new(bytes.Buffer)
	assert.NoError(t, render(&startResult{
		Success: true,
		ClusterConfig: &clusterConfig{
			WebConsoleURL: defaultWebConsoleURL,
			URL:           defaultAPIURL,
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

	var userShell string
	userShell, err := shell.GetShell("")
	if err != nil {
		if runtime.GOOS == "windows" {
			userShell, err = shell.GetShell("cmd")
		} else {
			userShell, err = shell.GetShell("bash")
		}
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedTemplate(userShell), out.String())
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
			WebConsoleURL: defaultWebConsoleURL,
			URL:           defaultAPIURL,
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

const unixTemplate = `Started the OpenShift cluster.

The server is accessible via web console at:
  https://console-openshift-console.apps-crc.testing

Log in as administrator:
  Username: kubeadmin
  Password: secret

Log in as user:
  Username: developer
  Password: developer

Use the 'oc' command line interface:
  $ eval $(crc oc-env)
  $ oc login -u developer https://api.crc.testing:6443
`

const powershellTemplate = `Started the OpenShift cluster.

The server is accessible via web console at:
  https://console-openshift-console.apps-crc.testing

Log in as administrator:
  Username: kubeadmin
  Password: secret

Log in as user:
  Username: developer
  Password: developer

Use the 'oc' command line interface:
  PS> & crc oc-env | Invoke-Expression
  PS> oc login -u developer https://api.crc.testing:6443
`

const cmdTemplate = `Started the OpenShift cluster.

The server is accessible via web console at:
  https://console-openshift-console.apps-crc.testing

Log in as administrator:
  Username: kubeadmin
  Password: secret

Log in as user:
  Username: developer
  Password: developer

Use the 'oc' command line interface:
  > @FOR /f "tokens=*" %i IN ('crc oc-env') DO @call %i
  > oc login -u developer https://api.crc.testing:6443
`

func expectedTemplate(shell string) string {
	if runtime.GOOS == "windows" {
		if shell == "powershell" {
			return powershellTemplate
		}
		return cmdTemplate
	}
	return unixTemplate
}
