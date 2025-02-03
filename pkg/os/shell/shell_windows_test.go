package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect_WhenUnknownShell_ThenDefaultToCmdShell(t *testing.T) {
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
					name: "crc.exe",
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
					name: "crc.exe",
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
					name: "crc.exe",
				},
				{
					name: "unknown.exe",
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

			// When
			shell, err := detect()

			// Then
			assert.NoError(t, err)
			assert.Equal(t, "cmd", shell)
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
					name: "crc.exe",
				},
				{
					name: "bash.exe",
				},
			},
			"bash",
		},
		{
			"powershell, then detect powershell",
			[]MockedProcess{
				{
					name: "crc.exe",
				},
				{
					name: "powershell.exe",
				},
			},
			"powershell",
		},
		{
			"cmd shell, then detect fish shell",
			[]MockedProcess{
				{
					name: "crc.exe",
				},
				{
					name: "cmd.exe",
				},
			},
			"cmd",
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

func TestSupportedShells(t *testing.T) {
	assert.Equal(t, []string{"cmd", "powershell", "wsl", "bash", "zsh", "fish"}, supportedShell)
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

func TestInspectProcessForRecentlyUsedShell(t *testing.T) {
	tests := []struct {
		name              string
		psCommandOutput   string
		expectedShellType string
	}{
		{
			"nothing provided, then return empty string",
			"",
			"",
		},
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
			// Given + When
			result := inspectProcessOutputForRecentlyUsedShell(tt.psCommandOutput)
			// Then
			if result != tt.expectedShellType {
				t.Errorf("%s inspectProcessOutputForRecentlyUsedShell() = %s; want %s", tt.name, result, tt.expectedShellType)
			}
		})
	}
}
