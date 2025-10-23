//go:build !windows

package shell

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownShell(t *testing.T) {
	tests := []struct {
		name              string
		processTree       []MockedProcess
		expectedShellType string
	}{
		{
			"failure to get process details for given pid",
			[]MockedProcess{},
			"",
		},
		{
			"failure while getting name of process",
			[]MockedProcess{
				{
					name: "crc",
				},
				{
					nameGetFails: true,
				},
			},
			"",
		},
		{
			"failure while getting ppid of process",
			[]MockedProcess{
				{
					name: "crc",
				},
				{
					parentGetFails: true,
				},
			},
			"",
		},
		{
			"failure when no shell process in process tree",
			[]MockedProcess{
				{
					name: "crc",
				},
				{
					name: "unknown",
				},
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			currentProcessSupplier = func() AbstractProcess {
				return createNewMockProcessTreeFrom(tt.processTree)
			}
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// When
			shell, err := detect()

			// Then
			assert.Equal(t, err, ErrUnknownShell)
			err = w.Close()
			assert.NoError(t, err)
			os.Stdout = originalStdout
			var buf bytes.Buffer
			nBytesRead, err := buf.ReadFrom(r)
			assert.NoError(t, err)
			assert.Greater(t, nBytesRead, int64(0))
			assert.Equal(t, "The default lines below are for a sh/bash shell, you can specify the shell you're using, with the --shell flag.\n\n", buf.String())
			assert.Empty(t, shell)
		})
	}
}

func TestDetect_GivenProcessTree_ThenReturnShellProcessWithCorrespondingParentPID(t *testing.T) {
	tests := []struct {
		name              string
		processTree       []MockedProcess
		expectedShellType string
	}{
		{
			"bash shell, then detect bash shell",
			[]MockedProcess{
				{
					name: "crc",
				},
				{
					name: "bash",
				},
			},
			"bash",
		},
		{
			"zsh shell, then detect zsh shell",
			[]MockedProcess{
				{
					name: "crc",
				},
				{
					name: "zsh",
				},
			},
			"zsh",
		},
		{
			"fish shell, then detect fish shell",
			[]MockedProcess{
				{
					name: "crc",
				},
				{
					name: "fish",
				},
			},
			"fish",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			currentProcessSupplier = func() AbstractProcess {
				return createNewMockProcessTreeFrom(tt.processTree)
			}
			// When
			shell, err := detect()

			// Then
			assert.Equal(t, tt.expectedShellType, shell)
			assert.NoError(t, err)
		})
	}
}

func TestGetCurrentProcess(t *testing.T) {
	// Given
	// When
	currentProcess := createCurrentProcess()

	// Then
	assert.NotNil(t, currentProcess)
	parentProcess, err := currentProcess.Parent()
	assert.NoError(t, err)
	assert.NotNil(t, parentProcess)
	currentProcessName, err := currentProcess.Name()
	assert.NoError(t, err)
	assert.Greater(t, len(currentProcessName), 0)
}
