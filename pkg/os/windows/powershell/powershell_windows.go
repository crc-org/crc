package powershell

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

var isAdminCmds = []string{
	"$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())",
	"$currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)",
}

func IsAdmin() bool {
	cmd := strings.Join(isAdminCmds, ";")
	stdOut, _, err := Execute(cmd)
	if err != nil {
		return false
	}
	if strings.TrimSpace(stdOut) == "False" {
		return false
	}

	return true
}

func Execute(args ...string) (string, string, error) {
	logging.Debugf("Running '%s'", strings.Join(args, " "))

	powershell, err := exec.LookPath("powershell.exe")
	if err != nil {
		return "", "", err
	}
	args = append([]string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "RemoteSigned", "-Command", "$ProgressPreference = 'SilentlyContinue';"}, args...)
	cmd := exec.Command(powershell, args...) // #nosec G204
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		logging.Debugf("Command failed: %v", err)
		logging.Debugf("stdout: %s", stdout.String())
		logging.Debugf("stderr: %s", stderr.String())
	}
	return stdout.String(), stderr.String(), err
}

func ExecuteAsAdmin(reason, cmd string) (string, string, error) {
	powershell, err := exec.LookPath("powershell.exe")
	if err != nil {
		return "", "", err
	}
	scriptContent := strings.Join(append(runAsCmds(powershell), cmd), "\n")

	tempDir, err := os.MkdirTemp("", "crcScripts")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(tempDir)

	// Write a temporary script
	/* Add UTF-8 BOM at the beginning of the script so that Windows
	 * correctly detects the file encoding
	 */
	filename := filepath.Join(tempDir, "runAsAdmin.ps1")

	// #nosec G306
	if err := os.WriteFile(filename, append([]byte{0xef, 0xbb, 0xbf}, []byte(scriptContent)...), 0600); err != nil {
		return "", "", err
	}

	logging.Infof("Will run as admin: %s", reason)

	return Execute(filename)
}

func runAsCmds(powershell string) []string {
	return []string{
		`$myWindowsID = [System.Security.Principal.WindowsIdentity]::GetCurrent();`,
		`$myWindowsPrincipal = New-Object System.Security.Principal.WindowsPrincipal($myWindowsID);`,
		`$adminRole = [System.Security.Principal.WindowsBuiltInRole]::Administrator;`,
		`if (-Not ($myWindowsPrincipal.IsInRole($adminRole))) {`,
		`  $procInfo = New-Object System.Diagnostics.ProcessStartInfo;`,
		`  $procInfo.FileName = "` + powershell + `"`,
		`  $procInfo.WindowStyle = [Diagnostics.ProcessWindowStyle]::Hidden`,
		`  $procInfo.Arguments = "-ExecutionPolicy RemoteSigned & '" + $script:MyInvocation.MyCommand.Path + "'"`,
		`  $procInfo.Verb = "runas";`,
		`  $p = New-Object System.Diagnostics.Process`,
		`  $p.StartInfo = $procInfo`,
		`  $p.Start() | Out-Null`,
		`  $p.WaitForExit()`,
		`  if ($p.ExitCode -ne 0) {`,
		`    throw "Unexpected failure";`,
		`  }`,
		`  Exit;`,
		`}`,
	}
}
