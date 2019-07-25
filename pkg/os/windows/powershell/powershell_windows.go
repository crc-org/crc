package powershell

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"os/exec"
)

var (
	runAsCmds = []string{
		`$myWindowsID = [System.Security.Principal.WindowsIdentity]::GetCurrent();`,
		`$myWindowsPrincipal = New-Object System.Security.Principal.WindowsPrincipal($myWindowsID);`,
		`$adminRole = [System.Security.Principal.WindowsBuiltInRole]::Administrator;`,
		`if (-Not ($myWindowsPrincipal.IsInRole($adminRole))) {`,
		`  $procInfo = New-Object System.Diagnostics.ProcessStartInfo;`,
		`  $procInfo.FileName = "` + locatePowerShell() + `"`,
		`  $procInfo.WindowStyle = [Diagnostics.ProcessWindowStyle]::Hidden`,
		`  $procInfo.Arguments = "& '" + $script:MyInvocation.MyCommand.Path + "'"`,
		`  $procInfo.Verb = "runas";`,
		`  [System.Diagnostics.Process]::Start($procInfo);`,
		`  Exit;`,
		`}`,
	}
	isAdminCmds = []string{
		"$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())",
		"$currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)",
	}
)

func locatePowerShell() string {
	ps, _ := exec.LookPath("powershell.exe")
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

func Execute(args ...string) (stdOut string, stdErr string, err error) {
	args = append([]string{"-NoProfile", "-NonInteractive", "-ExecutionPolicy", "RemoteSigned", "-Command"}, args...)
	cmd := exec.Command(locatePowerShell(), args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	stdOut, stdErr = stdout.String(), stderr.String()

	return
}

func ExecuteAsAdmin(cmd string) (stdOut string, stdErr string, err error) {
	scriptContent := strings.Join(append(runAsCmds, cmd), "\n")

	tempDir, _ := ioutil.TempDir("", "crcScripts")
	psFile, err := os.Create(filepath.Join(tempDir, "runAsAdmin.ps1"))
	if err != nil {
		return "", "", err
	}

	// Write a temporary script
	psFile.WriteString(scriptContent)
	psFile.Close()

	return Execute(psFile.Name())

	// TODO: cleanup the mess
}
