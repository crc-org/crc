package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apiClient "github.com/code-ready/crc/pkg/crc/api/client"
	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/stretchr/testify/assert"
)

func TestHTTPApi(t *testing.T) {
	fakeMachine := fakemachine.NewClient()
	config := setupNewInMemoryConfig()

	ts := httptest.NewServer(NewMux(config, fakeMachine, &mockLogger{}, &mockTelemetry{}))
	defer ts.Close()

	client := apiClient.New(http.DefaultClient, ts.URL)
	vr, err := client.Version()
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.VersionResult{
			CrcVersion:       version.GetCRCVersion(),
			OpenshiftVersion: version.GetBundleVersion(),
			CommitSha:        version.GetCommitSha(),
			Success:          true,
		},
		vr,
	)

	statusResult, err := client.Status()
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.ClusterStatusResult{
			Name:             "crc",
			CrcStatus:        "Running",
			OpenshiftStatus:  "Running",
			OpenshiftVersion: "4.5.1",
			DiskUse:          int64(10000000000),
			DiskSize:         int64(20000000000),
			Success:          true,
		},
		statusResult,
	)

	var startConfig = apiClient.StartConfig{}
	startResult, err := client.Start(startConfig)
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.StartResult{
			Name:           "crc",
			Status:         "",
			Error:          "",
			KubeletStarted: true,
			ClusterConfig: types.ClusterConfig{
				ClusterCACert: "MIIDODCCAiCgAwIBAgIIRVfCKNUa1wIwDQYJ",
				KubeConfig:    "/tmp/kubeconfig",
				KubeAdminPass: "foobar",
				ClusterAPI:    "https://foo.testing:6443",
				WebConsoleURL: "https://console.foo.testing:6443",
				ProxyConfig:   nil,
			},
		},
		startResult,
	)

	stopResult, err := client.Stop()
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.Result{
			Name:    "crc",
			Success: true,
			Error:   "",
		},
		stopResult,
	)

	deleteResult, err := client.Delete()
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.Result{
			Name:    "crc",
			Success: true,
			Error:   "",
		},
		deleteResult,
	)

	configGetResult, err := client.GetConfig([]string{"cpus"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.GetConfigResult{
			Error: "",
			Configs: map[string]interface{}{
				"cpus": float64(4),
			},
		},
		configGetResult,
	)

	configSetResult, err := client.SetConfig(apiClient.SetConfigRequest{
		Properties: map[string]interface{}{
			"cpus": float64(5),
		},
	})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.SetOrUnsetConfigResult{
			Error:      "",
			Properties: []string{"cpus"},
		},
		configSetResult,
	)
}

func TestTelemetry(t *testing.T) {
	fakeMachine := fakemachine.NewClient()
	config := setupNewInMemoryConfig()

	telemetry := &mockTelemetry{}
	ts := httptest.NewServer(NewMux(config, fakeMachine, &mockLogger{}, telemetry))
	defer ts.Close()

	client := apiClient.New(http.DefaultClient, ts.URL)

	_ = client.Telemetry("click start")
	_ = client.Telemetry("click stop")

	assert.Equal(t, []string{"click start", "click stop"}, telemetry.actions)
}
