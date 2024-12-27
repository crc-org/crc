package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockCommandRunner struct {
	commandName                string
	commandArgs                []string
	expectedOutputToReturn     string
	expectedErrMessageToReturn string
	expectedErrToReturn        error
}

func NewMockCommandRunner() *MockCommandRunner {
	return NewMockCommandRunnerWithOutputErr("", "", nil)
}

func NewMockCommandRunnerWithOutputErr(output string, errorMsg string, err error) *MockCommandRunner {
	return &MockCommandRunner{
		commandName:                "",
		commandArgs:                []string{},
		expectedOutputToReturn:     output,
		expectedErrMessageToReturn: errorMsg,
		expectedErrToReturn:        err,
	}
}

func (e *MockCommandRunner) Run(command string, args ...string) (string, string, error) {
	e.commandName = command
	e.commandArgs = args
	return e.expectedOutputToReturn, e.expectedErrMessageToReturn, e.expectedErrToReturn
}

func (e *MockCommandRunner) RunPrivate(command string, args ...string) (string, string, error) {
	e.commandName = command
	e.commandArgs = args
	return e.expectedOutputToReturn, e.expectedErrMessageToReturn, e.expectedErrToReturn
}

func (e *MockCommandRunner) RunPrivileged(_ string, cmdAndArgs ...string) (string, string, error) {
	e.commandArgs = cmdAndArgs
	return e.expectedOutputToReturn, e.expectedErrMessageToReturn, e.expectedErrToReturn
}

func TestGetPathEnvString(t *testing.T) {
	tests := []struct {
		name        string
		userShell   string
		path        string
		expectedStr string
	}{
		{"fish shell", "fish", "C:\\Users\\foo\\.crc\\bin\\oc", "contains /C/Users/foo/.crc/bin/oc $fish_user_paths; or set -U fish_user_paths /C/Users/foo/.crc/bin/oc $fish_user_paths"},
		{"powershell shell", "powershell", "C:\\Users\\foo\\oc.exe", "$Env:PATH = \"C:\\Users\\foo\\oc.exe;$Env:PATH\""},
		{"cmd shell", "cmd", "C:\\Users\\foo\\oc.exe", "SET PATH=C:\\Users\\foo\\oc.exe;%PATH%"},
		{"bash with windows path", "bash", "C:\\Users\\foo.exe", "export PATH=\"/C/Users/foo.exe:$PATH\""},
		{"unknown with windows path", "unknown", "C:\\Users\\foo.exe", "export PATH=\"C:\\Users\\foo.exe:$PATH\""},
		{"unknown shell with unix path", "unknown", "/home/foo/.crc/bin/oc", "export PATH=\"/home/foo/.crc/bin/oc:$PATH\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPathEnvString(tt.userShell, tt.path)
			if result != tt.expectedStr {
				t.Errorf("GetPathEnvString(%s, %s) = %s; want %s", tt.userShell, tt.path, result, tt.expectedStr)
			}
		})
	}
}

func TestConvertToLinuxStylePath(t *testing.T) {
	tests := []struct {
		name         string
		userShell    string
		path         string
		expectedPath string
	}{
		{"bash on windows, should convert", "bash", "C:\\Users\\foo\\.crc\\bin\\oc", "/C/Users/foo/.crc/bin/oc"},
		{"zsh on windows, should convert", "zsh", "C:\\Users\\foo\\.crc\\bin\\oc", "/C/Users/foo/.crc/bin/oc"},
		{"fish on windows, should convert", "fish", "C:\\Users\\foo\\.crc\\bin\\oc", "/C/Users/foo/.crc/bin/oc"},
		{"powershell on windows, should NOT convert", "powershell", "C:\\Users\\foo\\.crc\\bin\\oc", "C:\\Users\\foo\\.crc\\bin\\oc"},
		{"cmd on windows, should NOT convert", "cmd", "C:\\Users\\foo\\.crc\\bin\\oc", "C:\\Users\\foo\\.crc\\bin\\oc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToLinuxStylePath(tt.userShell, tt.path)
			if result != tt.expectedPath {
				t.Errorf("convertToLinuxStylePath(%s, %s) = %s; want %s", tt.userShell, tt.path, result, tt.expectedPath)
			}
		})
	}
}

func TestConvertToLinuxStylePath_WhenRunOnWSL_ThenExecuteWslPathBinary(t *testing.T) {
	// Given
	dir := t.TempDir()
	wslVersionFilePath := filepath.Join(dir, "version")
	wslVersionFile, err := os.Create(wslVersionFilePath)
	assert.NoError(t, err)
	defer func(wslVersionFile *os.File) {
		err := wslVersionFile.Close()
		assert.NoError(t, err)
	}(wslVersionFile)
	numberOfBytesWritten, err := wslVersionFile.WriteString("Linux version 5.15.167.4-microsoft-standard-WSL2 (root@f9c826d3017f) (gcc (GCC) 11.2.0, GNU ld (GNU Binutils) 2.37) #1 SMP Tue Nov 5 00:21:55 UTC 2024")
	assert.NoError(t, err)
	assert.Greater(t, numberOfBytesWritten, 0)
	WindowsSubsystemLinuxKernelMetadataFile = wslVersionFilePath
	mockCommandExecutor := NewMockCommandRunner()
	CommandRunner = mockCommandExecutor
	// When
	convertToLinuxStylePath("wsl", "C:\\Users\\foo\\.crc\\bin\\oc")
	// Then
	assert.Equal(t, "wsl", mockCommandExecutor.commandName)
	assert.Equal(t, []string{"-e", "bash", "-c", "wslpath -a 'C:\\Users\\foo\\.crc\\bin\\oc'"}, mockCommandExecutor.commandArgs)
}

func TestIsWindowsSubsystemLinux_whenInvalidKernelInfoFile_thenReturnFalse(t *testing.T) {
	// Given + When
	WindowsSubsystemLinuxKernelMetadataFile = "/i/dont/exist"
	// Then
	assert.Equal(t, false, IsWindowsSubsystemLinux())
}

func TestIsWindowsSubsystemLinux_whenValidKernelInfoFile_thenReturnTrue(t *testing.T) {
	tests := []struct {
		name               string
		versionFileContent string
		expectedResult     bool
	}{
		{
			"version file contains WSL and Microsoft keywords, then return true",
			"Linux version 5.15.167.4-microsoft-standard-WSL2 (root@f9c826d3017f) (gcc (GCC) 11.2.0, GNU ld (GNU Binutils) 2.37) #1 SMP Tue Nov 5 00:21:55 UTC 2024",
			true,
		},
		{
			"version file does NOT contain WSL and Microsoft keywords, then return false",
			"invalid",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			dir := t.TempDir()
			wslVersionFilePath := filepath.Join(dir, "version")
			wslVersionFile, err := os.Create(wslVersionFilePath)
			assert.NoError(t, err)
			defer func(wslVersionFile *os.File) {
				err := wslVersionFile.Close()
				assert.NoError(t, err)
				err = os.Remove(wslVersionFile.Name())
				assert.NoError(t, err)
			}(wslVersionFile)
			numberOfBytesWritten, err := wslVersionFile.WriteString(tt.versionFileContent)
			assert.NoError(t, err)
			assert.Greater(t, numberOfBytesWritten, 0)
			WindowsSubsystemLinuxKernelMetadataFile = wslVersionFilePath
			// When
			result := IsWindowsSubsystemLinux()

			// Then
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestConvertToWindowsSubsystemLinuxPath(t *testing.T) {
	// Given
	mockCommandExecutor := NewMockCommandRunner()
	CommandRunner = mockCommandExecutor

	// When
	convertToWindowsSubsystemLinuxPath("C:\\Users\\foo\\.crc\\bin\\oc")

	// Then
	assert.Equal(t, "wsl", mockCommandExecutor.commandName)
	assert.Equal(t, []string{"-e", "bash", "-c", "wslpath -a 'C:\\Users\\foo\\.crc\\bin\\oc'"}, mockCommandExecutor.commandArgs)
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
