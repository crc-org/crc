package shell

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "")

	shell, err := detect()

	assert.Contains(t, supportedShell, shell)
	assert.NoError(t, err)
}

func TestGetNameAndItsPpidOfCurrent(t *testing.T) {
	pid := os.Getpid()
	if pid < 0 || pid > math.MaxUint32 {
		assert.Fail(t, "integer overflow detected")
	}
	shell, shellppid, err := getNameAndItsPpid(uint32(pid))
	assert.Equal(t, "shell.test.exe", shell)
	ppid := os.Getppid()
	if ppid < 0 || ppid > math.MaxUint32 {
		assert.Fail(t, "integer overflow detected")
	}
	assert.Equal(t, uint32(ppid), shellppid)
	assert.NoError(t, err)
}

func TestGetNameAndItsPpidOfParent(t *testing.T) {
	pid := os.Getppid()
	if pid < 0 || pid > math.MaxUint32 {
		assert.Fail(t, "integer overflow detected")
	}
	shell, _, err := getNameAndItsPpid(uint32(pid))

	assert.Equal(t, "go.exe", shell)
	assert.NoError(t, err)
}

func TestSupportedShells(t *testing.T) {
	assert.Equal(t, []string{"cmd", "powershell", "bash", "zsh", "fish"}, supportedShell)
}

func TestShellType(t *testing.T) {
	tests := []struct {
		name              string
		userShell         string
		expectedShellType string
	}{
		{"git bash", "C:\\Program Files\\Git\\usr\\bin\\bash.exe", "bash"},
		{"windows subsystem for linux", "wsl.exe", "bash"},
		{"powershell", "powershell", "powershell"},
		{"cmd.exe", "cmd.exe", "cmd"},
		{"pwsh", "pwsh.exe", "powershell"},
		{"empty value", "", "cmd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			CommandRunner = NewMockCommandRunner()
			// When
			result := shellType(tt.userShell, "cmd")
			// Then
			if result != tt.expectedShellType {
				t.Errorf("shellType(%s) = %s; want %s", tt.userShell, result, tt.expectedShellType)
			}
		})
	}
}

func TestDetectShellInWindowsSubsystemLinux(t *testing.T) {
	// Given
	mockCommandExecutor := NewMockCommandRunner()
	CommandRunner = mockCommandExecutor

	// When
	shellType("wsl.exe", "cmd")

	// Then
	assert.Equal(t, "wsl", mockCommandExecutor.commandName)
	assert.Equal(t, []string{"-e", "bash", "-c", "ps -ao pid=,comm="}, mockCommandExecutor.commandArgs)
}
