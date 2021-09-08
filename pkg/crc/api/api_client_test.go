package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	apiClient "github.com/code-ready/crc/pkg/crc/api/client"
	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/stretchr/testify/assert"
)

type testClient struct {
	*apiClient.Client
	config     crcConfig.Storage
	httpServer *httptest.Server
}

func (client *testClient) Close() {
	client.httpServer.Close()
}

func newTestClient() *testClient {
	fakeMachine := fakemachine.NewClient()
	config := setupNewInMemoryConfig()

	ts := httptest.NewServer(NewMux(config, fakeMachine, &mockLogger{}, &mockTelemetry{}))

	return &testClient{
		apiClient.New(http.DefaultClient, ts.URL),
		config,
		ts,
	}
}

func TestVersion(t *testing.T) {
	client := newTestClient()
	defer client.Close()
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
}

func TestStatus(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	statusResult, err := client.Status()
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.ClusterStatusResult{
			CrcStatus:        "Running",
			OpenshiftStatus:  "Running",
			OpenshiftVersion: "4.5.1",
			DiskUse:          int64(10000000000),
			DiskSize:         int64(20000000000),
			Success:          true,
		},
		statusResult,
	)
}

func TestStart(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	startConfig := apiClient.StartConfig{}
	startResult, err := client.Start(startConfig)
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.StartResult{
			Status:         "",
			Error:          "",
			KubeletStarted: true,
			Success:        true,
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
}

func TestSetup(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	stopResult, err := client.Stop()
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.Result{
			Success: true,
			Error:   "",
		},
		stopResult,
	)
}

func TestDelete(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	deleteResult, err := client.Delete()
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.Result{
			Success: true,
			Error:   "",
		},
		deleteResult,
	)
}

func TestConfigGet(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	configGetResult, err := client.GetConfig([]string{"cpus"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.GetConfigResult{
			Success: true,
			Error:   "",
			Configs: map[string]interface{}{
				"cpus": float64(4),
			},
		},
		configGetResult,
	)
}

func TestConfigSet(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	configSetResult, err := client.SetConfig(apiClient.SetConfigRequest{
		Properties: map[string]interface{}{
			"cpus": float64(5),
		},
	})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.SetOrUnsetConfigResult{
			Success:    true,
			Error:      "",
			Properties: []string{"cpus"},
		},
		configSetResult,
	)
	// Get the cpus config again after setting it to make sure it is set
	// properly.
	configGetAfterSetResult, err := client.GetConfig([]string{"cpus"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.GetConfigResult{
			Success: true,
			Error:   "",
			Configs: map[string]interface{}{
				"cpus": float64(5),
			},
		},
		configGetAfterSetResult,
	)
}

func TestConfigUnset(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	configUnsetResult, err := client.UnsetConfig([]string{"cpus"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.SetOrUnsetConfigResult{
			Success:    true,
			Error:      "",
			Properties: []string{"cpus"},
		},
		configUnsetResult,
	)
}

func TestConfigGetAll(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	allConfigGetResult, err := client.GetConfig(nil)
	assert.NoError(t, err)
	configs := make(map[string]interface{})
	for k, v := range client.config.AllConfigs() {
		// This is required because of https://pkg.go.dev/encoding/json#Unmarshal
		// Unmarshal stores float64 for JSON numbers in case of interface.
		switch v := v.Value.(type) {
		case int:
			configs[k] = float64(v)
		default:
			configs[k] = v
		}
	}
	assert.Equal(
		t,
		apiClient.GetConfigResult{
			Success: true,
			Error:   "",
			Configs: configs,
		},
		allConfigGetResult,
	)
}

func TestConfigGetMultiple(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	// Get result of making query for multiple property
	configGetMultiplePropertyResult, err := client.GetConfig([]string{"cpus", "memory"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.GetConfigResult{
			Success: true,
			Error:   "",
			Configs: map[string]interface{}{
				"cpus":   float64(4),
				"memory": float64(9216),
			},
		},
		configGetMultiplePropertyResult,
	)
}

func TestConfigGetEscaped(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	// Get result of special query for multiple properties
	configGetSpecialPropertyResult, err := client.GetConfig([]string{"a&a", "b&&&b"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.GetConfigResult{
			Success: true,
			Error:   "",
			Configs: map[string]interface{}{
				"a&a":   "foo",
				"b&&&b": "bar",
			},
		},
		configGetSpecialPropertyResult,
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

func TestPullSecret(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-pull-secret")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	fakeMachine := fakemachine.NewClient()
	config := setupNewInMemoryConfig()

	ts := httptest.NewServer(NewMux(config, fakeMachine, &mockLogger{}, &mockTelemetry{}))
	defer ts.Close()

	client := apiClient.New(http.DefaultClient, ts.URL)

	defined, err := client.IsPullSecretDefined()
	assert.NoError(t, err)
	assert.False(t, defined)

	pullSecretFile := filepath.Join(dir, "pull-secret.json")
	assert.NoError(t, ioutil.WriteFile(pullSecretFile, []byte(constants.OkdPullSecret), 0600))
	_, err = config.Set(crcConfig.PullSecretFile, pullSecretFile)
	assert.NoError(t, err)

	defined, err = client.IsPullSecretDefined()
	assert.NoError(t, err)
	assert.False(t, defined)

	assert.Error(t, client.SetPullSecret("{}")) // invalid
}
