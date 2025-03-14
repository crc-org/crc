package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mocks "github.com/crc-org/crc/v2/test/mocks/api"

	apiClient "github.com/crc-org/crc/v2/pkg/crc/api/client"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/preset"

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
		RAMUse:           8_000_000_000,
		RAMSize:          10_000_000_000,
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
	cacheDir := t.TempDir()

	client := setUpClient(t)

	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, "", false))

	expected := `CRC VM:          Running
OpenShift:       Running (v4.5.1)
RAM Usage:       8GB of 10GB
Disk Usage:      10GB of 20GB (Inside the CRC VM)
Cache Usage:     10kB
Cache Directory: %s
`
	assert.Equal(t, fmt.Sprintf(expected, cacheDir), out.String())
}

func TestStatusWithoutPodman(t *testing.T) {
	cacheDir := t.TempDir()

	client := mocks.NewClient(t)
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	client.On("Status").Return(apiClient.ClusterStatusResult{
		CrcStatus:        string(state.Running),
		OpenshiftStatus:  string(types.OpenshiftRunning),
		OpenshiftVersion: "4.5.1",
		RAMUse:           15_000_000_000,
		RAMSize:          20_000_000_000,
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
		Preset:           preset.OpenShift,
	}, nil)

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, "", false))

	expected := `CRC VM:          Running
OpenShift:       Running (v4.5.1)
RAM Usage:       15GB of 20GB
Disk Usage:      10GB of 20GB (Inside the CRC VM)
Cache Usage:     10kB
Cache Directory: %s
`
	assert.Equal(t, fmt.Sprintf(expected, cacheDir), out.String())
}

func TestJsonStatus(t *testing.T) {
	cacheDir := t.TempDir()

	client := setUpClient(t)

	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, jsonFormat, false))

	expected := `{
  "success": true,
  "crcStatus": "Running",
  "openshiftStatus": "Running",
  "openshiftVersion": "4.5.1",
  "diskUsage": 10000000000,
  "diskSize": 20000000000,
  "cacheUsage": 10000,
  "cacheDir": "%s",
  "ramSize": 10000000000,
  "ramUsage": 8000000000,
  "preset": "openshift"
}
`
	assert.Equal(t, fmt.Sprintf(expected, strings.ReplaceAll(cacheDir, `\`, `\\`)), out.String())
}

func TestPlainStatusWithError(t *testing.T) {
	cacheDir := t.TempDir()

	client := setUpFailingClient(t)

	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.EqualError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, "", false), "broken")
	assert.Equal(t, "", out.String())
}

func TestJsonStatusWithError(t *testing.T) {
	cacheDir := t.TempDir()

	client := setUpFailingClient(t)

	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

	out := new(bytes.Buffer)
	assert.NoError(t, runStatus(out, &daemonclient.Client{
		APIClient: client,
	}, cacheDir, jsonFormat, false))

	expected := `{
  "success": false,
  "error": "broken",
  "preset": ""
}
`
	assert.Equal(t, expected, out.String())
}

func TestStatusWithMemoryPodman(t *testing.T) {
	cacheDir := t.TempDir()

	client := mocks.NewClient(t)
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))

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
	}, cacheDir, "", false))

	expected := `CRC VM:          Running
OpenShift:       Running (v4.5.1)
RAM Usage:       900kB of 1MB
Disk Usage:      10GB of 20GB (Inside the CRC VM)
Cache Usage:     10kB
Cache Directory: %s
`
	assert.Equal(t, fmt.Sprintf(expected, cacheDir), out.String())
}

func TestCrcStatusShouldLogInformationForConfiguredPresets(t *testing.T) {
	tests := []struct {
		name      string
		preset    preset.Preset
		statusLog string
	}{
		{"OpenShift preset should log OpenShift", preset.OpenShift, "OpenShift:       Running (v4.5.1)"},
		{"OKD preset should log OKD", preset.OKD, "OKD:             Running (v4.5.1)"},
		{"MicroShift preset should log MicroShift", preset.Microshift, "MicroShift:              Running (v4.5.1)"},
		{"Unknown preset should log OpenShift", preset.ParsePreset("unknown"), "OpenShift:       Running (v4.5.1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			cacheDir := t.TempDir()
			out := new(bytes.Buffer)
			client := mocks.NewClient(t)
			require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "crc.qcow2"), make([]byte, 10000), 0600))
			client.On("Status").Return(apiClient.ClusterStatusResult{
				CrcStatus:        string(state.Running),
				OpenshiftStatus:  string(types.OpenshiftRunning),
				OpenshiftVersion: "4.5.1",
				Preset:           tt.preset,
			}, nil)

			// When
			err := runStatus(out, &daemonclient.Client{
				APIClient: client,
			}, cacheDir, "", false)

			// Then
			assert.NoError(t, err)
			assert.Contains(t, out.String(), tt.statusLog)
		})
	}
}
