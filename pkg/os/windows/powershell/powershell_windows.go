package powershell

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"os/exec"

	"github.com/code-ready/crc/pkg/crc/logging"
)

var (
	runAsCmds = []string{
		`$myWindowsID = [System.Security.Principal.WindowsIdentity]::GetCurrent();`,
		`$myWindowsPrincipal = New-Object System.Security.Principal.WindowsPrincipal($myWindowsID);`,
		`$adminRole = [System.Security.Principal.WindowsBuiltInRole]::Administrator;`,
		`if (-Not ($myWindowsPrincipal.IsInRole($adminRole))) {`,
		`  $procInfo = New-Object System.Diagnostics.ProcessStartInfo;`,
		`  $procInfo.FileName = "` + LocatePowerShell() + `"`,
		`  $procInfo.WindowStyle = [Diagnostics.ProcessWindowStyle]::Hidden`,
		`  $procInfo.Arguments = "-ExecutionPolicy RemoteSigned & '" + $script:MyInvocation.MyCommand.Path + "'"`,
		`  $procInfo.Verb = "runas";`,
		`  [System.Diagnostics.Process]::Start($procInfo);`,
		`  Exit;`,
		`}`,
	}
	isAdminCmds = []string{
		"$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())",
		"$currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)",
	}

	poshStdLocation = filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
)

func LocatePowerShell() string {
	ps, err := exec.LookPath("powershell.exe")
	if err != nil {
		logging.Debugf("Could not find powershell.exe on path %s", err.Error())
		logging.Debug("Falling back to ", poshStdLocation)
		return poshStdLocation
	}
	return ps
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

	args = append([]string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "RemoteSigned", "-Command"}, args...)
	cmd := exec.Command(LocatePowerShell(), args...) // #nosec G204
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logging.Debugf("Command failed: %v", err)
		logging.Debugf("stdout: %s", stdout.String())
		logging.Debugf("stderr: %s", stderr.String())
	}
	return stdout.String(), stderr.String(), err
}

func ExecuteAsAdmin(reason, cmd string) (string, string, error) {
	scriptContent := strings.Join(append(runAsCmds, cmd), "\n")

	tempDir, _ := ioutil.TempDir("", "crcScripts")
	psFile, err := os.Create(filepath.Join(tempDir, "runAsAdmin.ps1"))
	if err != nil {
		return "", "", err
	}

	// Write a temporary script
	/* Add UTF-8 BOM at the beginning of the script so that Windows
	 * correctly detects the file encoding
	 */
	_, err = psFile.Write([]byte{0xef, 0xbb, 0xbf})
	if err != nil {
		return "", "", err
	}
	_, err = psFile.WriteString(scriptContent)
	if err != nil {
		return "", "", err
	}
	psFile.Close()
	logging.Infof("Will run as admin: %s", reason)

	return Execute(psFile.Name())

	// TODO: cleanup the mess
}
