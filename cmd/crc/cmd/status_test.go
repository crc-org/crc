package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mocks "github.com/code-ready/crc/test/mocks/api"

	apiClient "github.com/code-ready/crc/pkg/crc/api/client"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	"github.com/code-ready/crc/pkg/crc/machine/state"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/preset"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setUpClient(t *testing.T) *mocks.Client {
	client := mocks.NewClient(t)

	client.On("Status").Return(apiClient.ClusterStatusResult{
		CrcStatus:        string(state.Running),
		OpenshiftStatus:  string(types.OpenshiftRunning),
		OpenshiftVersion: "4.5.1",
		PodmanVersion:    "3.3.1",
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
		Preset:           preset.OpenShift,
	}, nil)

	return client
}

func setUpFailingClient(t *testing.T) *mocks.Client {
	client := mocks.NewClient(t)

	client.On("Status").Return(apiClient.ClusterStatusResult{}, errors.New("broken"))

	return client
}

func TestPlainStatus(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	client := setUpClient(t)

	require.NoError(t, ioutil.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, ""))

	expected := `CRC VM:          Running
OpenShift:       Running (v4.5.1)
Podman:          3.3.1
RAM Usage:       0B of 0B
Disk Usage:      10GB of 20GB (Inside the CRC VM)
Cache Usage:     10kB
Cache Directory: %s
`
	assert.Equal(t, fmt.Sprintf(expected, cacheDir), out.String())
}

func TestStatusWithoutPodman(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	client := mocks.NewClient(t)
	require.NoError(t, ioutil.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	client.On("Status").Return(apiClient.ClusterStatusResult{
		CrcStatus:        string(state.Running),
		OpenshiftStatus:  string(types.OpenshiftRunning),
		OpenshiftVersion: "4.5.1",
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
		Preset:           preset.OpenShift,
	}, nil)

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, ""))

	expected := `CRC VM:          Running
OpenShift:       Running (v4.5.1)
RAM Usage:       0B of 0B
Disk Usage:      10GB of 20GB (Inside the CRC VM)
Cache Usage:     10kB
Cache Directory: %s
`
	assert.Equal(t, fmt.Sprintf(expected, cacheDir), out.String())
}

func TestJsonStatus(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	client := setUpClient(t)

	require.NoError(t, ioutil.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, jsonFormat))

	expected := `{
  "success": true,
  "crcStatus": "Running",
  "openshiftStatus": "Running",
  "openshiftVersion": "4.5.1",
  "podmanVersion": "3.3.1",
  "diskUsage": 10000000000,
  "diskSize": 20000000000,
  "cacheUsage": 10000,
  "cacheDir": "%s",
  "preset": "openshift"
}
`
	assert.Equal(t, fmt.Sprintf(expected, strings.ReplaceAll(cacheDir, `\`, `\\`)), out.String())
}

func TestPlainStatusWithError(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	client := setUpFailingClient(t)

	require.NoError(t, ioutil.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.EqualError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, ""), "broken")
	assert.Equal(t, "", out.String())
}

func TestJsonStatusWithError(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	client := setUpFailingClient(t)

	require.NoError(t, ioutil.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, jsonFormat))

	expected := `{
  "success": false,
  "error": "broken",
  "preset": ""
}
`
	assert.Equal(t, expected, out.String())
}

func TestStatusWithMemoryPodman(t *testing.T) {
	cacheDir, err := ioutil.TempDir("", "cache")
	require.NoError(t, err)
	defer os.RemoveAll(cacheDir)

	client := mocks.NewClient(t)
	require.NoError(t, ioutil.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	client.On("Status").Return(apiClient.ClusterStatusResult{
		CrcStatus:        string(state.Running),
		OpenshiftStatus:  string(types.OpenshiftRunning),
		OpenshiftVersion: "4.5.1",
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
		RAMSize:          1_000_000,
		RAMUse:           900_000,
		Preset:           preset.OpenShift,
	}, nil)

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, ""))

	expected := `CRC VM:          Running
OpenShift:       Running (v4.5.1)
RAM Usage:       900kB of 1MB
Disk Usage:      10GB of 20GB (Inside the CRC VM)
Cache Usage:     10kB
Cache Directory: %s
`
	assert.Equal(t, fmt.Sprintf(expected, cacheDir), out.String())
}
