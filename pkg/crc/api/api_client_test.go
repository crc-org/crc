package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	apiClient "github.com/crc-org/crc/v2/pkg/crc/api/client"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/fakemachine"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/version"
	"github.com/stretchr/testify/assert"
)

type testClient struct {
	apiClient.Client
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
			OpenshiftVersion: version.GetBundleVersion(preset.OpenShift),
			CommitSha:        version.GetCommitSha(),
			PodmanVersion:    version.GetBundleVersion(preset.Podman),
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
			PodmanVersion:    "3.3.1",
			DiskUse:          int64(10000000000),
			DiskSize:         int64(20000000000),
			RAMUse:           int64(1000),
			RAMSize:          int64(2000),
			Preset:           preset.OpenShift,
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
			KubeletStarted: true,
			ClusterConfig: types.ClusterConfig{
				ClusterType:   "openshift",
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
	err := client.Stop()
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	err := client.Delete()
	assert.NoError(t, err)
}

func TestConfigGet(t *testing.T) {
	client := newTestClient()
	defer client.Close()
	configGetResult, err := client.GetConfig([]string{"cpus"})
	assert.NoError(t, err)
	assert.Equal(
		t,
		apiClient.GetConfigResult{
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
		// since we filter out secret configs at the config api handler level
		// we need exclude them from AllConfigs
		if v.IsSecret {
			continue
		}
		// This is required because of https://pkg.go.dev/encoding/json#Unmarshal
		// Unmarshal stores float64 for JSON numbers in case of interface.
		switch v := v.Value.(type) {
		case int:
			configs[k] = float64(v)
		// when config of type Path is converted to JSON it is converted to string
		case crcConfig.Path:
			configs[k] = string(v)
		default:
			configs[k] = v
		}
	}
	assert.Equal(
		t,
		apiClient.GetConfigResult{
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
	dir := t.TempDir()

	fakeMachine := fakemachine.NewClient()
	config := setupNewInMemoryConfig()

	ts := httptest.NewServer(NewMux(config, fakeMachine, &mockLogger{}, &mockTelemetry{}))
	defer ts.Close()

	client := apiClient.New(http.DefaultClient, ts.URL)

	defined, err := client.IsPullSecretDefined()
	assert.NoError(t, err)
	assert.False(t, defined)

	pullSecretFile := filepath.Join(dir, "pull-secret.json")
	assert.NoError(t, os.WriteFile(pullSecretFile, []byte(constants.OkdPullSecret), 0600))
	_, err = config.Set(crcConfig.PullSecretFile, pullSecretFile)
	assert.NoError(t, err)

	defined, err = client.IsPullSecretDefined()
	assert.NoError(t, err)
	assert.True(t, defined)

	assert.Error(t, client.SetPullSecret("{}")) // invalid
}
