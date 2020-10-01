// +build !windows

package api

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/fakemachine"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApi(t *testing.T) {
	dir, err := ioutil.TempDir("", "api")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	socket := filepath.Join(dir, "api.sock")
	listener, err := net.Listen("unix", socket)
	require.NoError(t, err)

	client := fakemachine.NewClient()
	api, err := createAPIServerWithListener(listener, client, config.New(config.NewEmptyInMemoryStorage()))
	require.NoError(t, err)
	go api.Serve()

	tt := []struct {
		command       string
		clientFailing bool
		args          json.RawMessage
		expected      map[string]interface{}
	}{
		{
			command: "version",
			expected: map[string]interface{}{
				"CrcVersion":       version.GetCRCVersion(),
				"CommitSha":        version.GetCommitSha(),
				"OpenshiftVersion": version.GetBundleVersion(),
				"Success":          true,
			},
		},
		{
			command: "status",
			expected: map[string]interface{}{
				"Name":             "crc",
				"CrcStatus":        "Running",
				"OpenshiftStatus":  "Running",
				"OpenshiftVersion": "4.5.1",
				"DiskUse":          float64(10000000000),
				"DiskSize":         float64(20000000000),
				"Error":            "",
				"Success":          true,
			},
		},
		{
			command:       "status",
			clientFailing: true,
			expected: map[string]interface{}{
				"Name":             "",
				"CrcStatus":        "",
				"OpenshiftStatus":  "",
				"OpenshiftVersion": "",
				"DiskUse":          float64(0),
				"DiskSize":         float64(0),
				"Error":            "broken",
				"Success":          false,
			},
		},
		{
			command:       "start",
			clientFailing: true,
			args:          json.RawMessage(`{"pull-secret":"/Users/fake/pull-secret"}`),
			expected: map[string]interface{}{
				"Name":           "crc",
				"Status":         "",
				"Error":          "Incorrect arguments given: json: unknown field \"pull-secret\"",
				"KubeletStarted": false,
				"ClusterConfig": map[string]interface{}{
					"KubeConfig":    "",
					"KubeAdminPass": "",
					"ClusterAPI":    "",
					"WebConsoleURL": "",
					"ProxyConfig":   nil,
				},
			},
		},
		{
			command: "start",
			args:    json.RawMessage(`{"pullSecretFile":"/Users/fake/pull-secret"}`),
			expected: map[string]interface{}{
				"Name":           "crc",
				"Status":         "",
				"Error":          "",
				"KubeletStarted": true,
				"ClusterConfig": map[string]interface{}{
					"KubeConfig":    "/tmp/kubeconfig",
					"KubeAdminPass": "foobar",
					"ClusterAPI":    "https://foo.testing:6443",
					"WebConsoleURL": "https://console.foo.testing:6443",
					"ProxyConfig":   nil,
				},
			},
		},
	}
	for _, test := range tt {
		client.Failing = test.clientFailing
		client, err := net.Dial("unix", socket)
		require.NoError(t, err)

		jsonReq, err := json.Marshal(commandRequest{
			Command: test.command,
			Args:    test.args,
		})
		assert.NoError(t, err)
		_, err = client.Write(jsonReq)
		assert.NoError(t, err)

		payload := make([]byte, 1024)
		n, err := client.Read(payload)
		assert.NoError(t, err)

		var res map[string]interface{}
		assert.NoError(t, json.Unmarshal(payload[0:n], &res))
		assert.Equal(t, test.expected, res)
	}
}

func TestSetconfigApi(t *testing.T) {
	// setup viper
	err := constants.EnsureBaseDirExists()
	assert.NoError(t, err)
	config := config.New(config.NewEmptyInMemoryStorage())
	assert.NoError(t, err)
	cmdConfig.RegisterSettings(config)

	socket, cleanup := setupAPIServer(t)
	client, err := net.Dial("unix", socket)
	require.NoError(t, err)
	defer cleanup()

	jsonReq, err := json.Marshal(commandRequest{
		Command: "setconfig",
		Args:    json.RawMessage(`{"properties":{"cpus":"5"}}`),
	})
	assert.NoError(t, err)
	_, err = client.Write(jsonReq)
	assert.NoError(t, err)

	payload := make([]byte, 1024)
	n, err := client.Read(payload)
	assert.NoError(t, err)

	var setconfigRes setOrUnsetConfigResult
	assert.NoError(t, json.Unmarshal(payload[:n], &setconfigRes))
	assert.Equal(t, setOrUnsetConfigResult{
		Error:      "",
		Properties: []string{"cpus"},
	}, setconfigRes)
}

func TestGetconfigApi(t *testing.T) {
	// setup viper
	err := constants.EnsureBaseDirExists()
	assert.NoError(t, err)
	config := config.New(config.NewEmptyInMemoryStorage())
	assert.NoError(t, err)
	cmdConfig.RegisterSettings(config)

	socket, cleanup := setupAPIServer(t)
	client, err := net.Dial("unix", socket)
	require.NoError(t, err)
	defer cleanup()

	jsonReq, err := json.Marshal(commandRequest{
		Command: "getconfig",
		Args:    json.RawMessage(`{"properties":["cpus"]}`),
	})
	assert.NoError(t, err)
	_, err = client.Write(jsonReq)
	assert.NoError(t, err)

	payload := make([]byte, 1024)
	n, err := client.Read(payload)
	assert.NoError(t, err)

	var getconfigRes getConfigResult
	assert.NoError(t, json.Unmarshal(payload[:n], &getconfigRes))

	configs := make(map[string]interface{})
	configs["cpus"] = 4.0

	assert.Equal(t, getConfigResult{
		Error:   "",
		Configs: configs,
	}, getconfigRes)
}

func setupAPIServer(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "api")
	require.NoError(t, err)

	socket := filepath.Join(dir, "api.sock")
	listener, err := net.Listen("unix", socket)
	require.NoError(t, err)

	client := fakemachine.NewClient()
	cfg := config.New(config.NewEmptyInMemoryStorage())
	cmdConfig.RegisterSettings(cfg)

	api, err := createAPIServerWithListener(listener, client, cfg)
	require.NoError(t, err)
	go api.Serve()

	return socket, func() { os.RemoveAll(dir) }
}
