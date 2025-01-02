//go:build !windows
// +build !windows

package shell

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownShell(t *testing.T) {
	// Given
	mockCommandExecutor := NewMockCommandRunnerWithOutputErr("", "", nil)
	CommandRunner = mockCommandExecutor
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
	assert.Equal(t, "ps", mockCommandExecutor.commandName)
	assert.Equal(t, []string{"-o", "pid=,comm="}, mockCommandExecutor.commandArgs)
	assert.Empty(t, shell)
}

func TestDetect_GivenPsOutputContainsShell_ThenReturnShellProcessWithMostRecentPid(t *testing.T) {
	tests := []struct {
		name              string
		psCommandOutput   string
		expectedShellType string
	}{
		{
			"bash shell, then detect bash shell",
			"  31162 ps\n  435 bash",
			"bash",
		},
		{
			"zsh shell, then detect zsh shell",
			"  31259 ps\n31253 zsh",
			"zsh",
		},
		{
			"fish shell, then detect fish shell",
			"  31372 ps\n  31352 fish",
			"fish",
		},
		{"bash and zsh shell, then detect zsh with more recent process id",
			"  31259 ps\n  31253 zsh\n  435 bash",
			"zsh",
		},
		{"bash and fish shell, then detect fish shell with more recent process id",
			"    31372 ps\n  31352 fish\n  435 bash",
			"fish",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			mockCommandExecutor := NewMockCommandRunnerWithOutputErr(tt.psCommandOutput, "", nil)
			CommandRunner = mockCommandExecutor

			// When
			shell, err := detect()

			// Then
			assert.Equal(t, "ps", mockCommandExecutor.commandName)
			assert.Equal(t, []string{"-o", "pid=,comm="}, mockCommandExecutor.commandArgs)
			assert.Equal(t, tt.expectedShellType, shell)
			assert.NoError(t, err)
		})
	}
}
