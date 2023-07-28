package preflight

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/pkg/crc/constants"
	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/crc-org/crc/pkg/crc/version"
	crcos "github.com/crc-org/crc/pkg/os"
	"github.com/crc-org/crc/pkg/os/windows/powershell"
)

var (
	// https://docs.microsoft.com/en-us/windows/win32/taskschd/daily-trigger-example--xml-
	daemonTaskTemplate = `<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.3" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Description>Run crc daemon as a task</Description>
    <Version>%s</Version>
  </RegistrationInfo>
  <Settings>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <Hidden>true</Hidden>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <IdleSettings>
      <Duration>PT10M</Duration>
      <WaitTimeout>PT1H</WaitTimeout>
      <StopOnIdleEnd>true</StopOnIdleEnd>
      <RestartOnIdle>false</RestartOnIdle>
    </IdleSettings>
    <UseUnifiedSchedulingEngine>true</UseUnifiedSchedulingEngine>
    <ExecutionTimeLimit>PT0S</ExecutionTimeLimit>
  </Settings>
  <Triggers>
    <LogonTrigger>
      <UserId>%s</UserId>
    </LogonTrigger>
  </Triggers>
  <Actions Context="Author">
    <Exec>
      <Command>powershell.exe</Command>
      <Arguments>-WindowStyle Hidden -Command %s</Arguments>
    </Exec>
  </Actions>
</Task>
`
	errOlderVersion = fmt.Errorf("expected %s task to be on version '%s'", constants.DaemonTaskName, version.GetCRCVersion())

	daemonPoshScriptTemplate = `# Following script is from https://stackoverflow.com/a/74976541
function Hide-ConsoleWindow() {
  $ShowWindowAsyncCode = '[DllImport("user32.dll")] public static extern bool ShowWindowAsync(IntPtr hWnd, int nCmdShow);'
  $ShowWindowAsync = Add-Type -MemberDefinition $ShowWindowAsyncCode -name Win32ShowWindowAsync -namespace Win32Functions -PassThru

  $hwnd = (Get-Process -PID $pid).MainWindowHandle
  if ($hwnd -ne [System.IntPtr]::Zero) {
    # When you got HWND of the console window:
    # (It would appear that Windows Console Host is the default terminal application)
    $ShowWindowAsync::ShowWindowAsync($hwnd, 0)
  } else {
    # When you failed to get HWND of the console window:
    # (It would appear that Windows Terminal is the default terminal application)

    # Mark the current console window with a unique string.
    $UniqueWindowTitle = New-Guid
    $Host.UI.RawUI.WindowTitle = $UniqueWindowTitle
    $StringBuilder = New-Object System.Text.StringBuilder 1024

    # Search the process that has the window title generated above.
    $TerminalProcess = (Get-Process | Where-Object { $_.MainWindowTitle -eq $UniqueWindowTitle })
    # Get the window handle of the terminal process.
    # Note that GetConsoleWindow() in Win32 API returns the HWND of
    # powershell.exe itself rather than the terminal process.
    # When you call ShowWindowAsync(HWND, 0) with the HWND from GetConsoleWindow(),
    # the Windows Terminal window will be just minimized rather than hidden.
    $hwnd = $TerminalProcess.MainWindowHandle
    if ($hwnd -ne [System.IntPtr]::Zero) {
      $ShowWindowAsync::ShowWindowAsync($hwnd, 0)
    } else {
      Write-Host "Failed to hide the console window."
    }
  }
}
Hide-ConsoleWindow
& "%s" %s`
)

func genDaemonTaskInstallTemplate(crcVersion, userName, daemonCommand string) (string, error) {
	var escapedName bytes.Buffer
	if err := xml.EscapeText(&escapedName, []byte(daemonCommand)); err != nil {
		return "", err
	}

	return fmt.Sprintf(daemonTaskTemplate,
		crcVersion,
		userName,
		escapedName.String(),
	), nil
}

func checkIfDaemonTaskInstalled() error {
	_, stderr, err := powershell.Execute("Get-ScheduledTask", "-TaskName", constants.DaemonTaskName)
	if err != nil {
		logging.Debugf("%s task is not installed: %v : %s", constants.DaemonTaskName, err, stderr)
		return err
	}
	if err := checkIfOlderTask(); err != nil {
		return err
	}
	return nil
}

func fixDaemonTaskInstalled() error {
	// Remove older task if exist
	if err := removeDaemonTask(); err != nil {
		return err
	}
	binPathWithArgs := fmt.Sprintf("& '%s'", daemonPoshScriptPath)
	// Get current user along with domain
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	taskContent, err := genDaemonTaskInstallTemplate(
		version.GetCRCVersion(),
		u.Username,
		binPathWithArgs,
	)
	if err != nil {
		return err
	}

	if _, stderr, err := powershell.Execute("Register-ScheduledTask", "-Xml", fmt.Sprintf(`'%s'`, taskContent), "-TaskName", constants.DaemonTaskName); err != nil {
		return fmt.Errorf("failed to register %s task, %v: %s", constants.DaemonTaskName, err, stderr)
	}

	return nil
}

func removeDaemonTask() error {
	// Return nil if the task does not exist
	_, stderr, err := powershell.Execute("Get-ScheduledTask", "-TaskName", constants.DaemonTaskName)
	if err != nil {
		logging.Debugf("%s task is not installed: %v : %s", constants.DaemonTaskName, err, stderr)
		return nil
	}
	if err := checkIfDaemonTaskRunning(); err == nil {
		_, stderr, err := powershell.Execute("Stop-ScheduledTask", "-TaskName", constants.DaemonTaskName)
		if err != nil {
			logging.Debugf("unable to stop the %s task: %v : %s", constants.DaemonTaskName, err, stderr)
			return err
		}
	}
	if err := checkIfDaemonTaskInstalled(); err == nil || errors.Is(err, errOlderVersion) {
		_, stderr, err := powershell.Execute("Unregister-ScheduledTask", "-TaskName", constants.DaemonTaskName, "-Confirm:$false")
		if err != nil {
			logging.Debugf("unable to unregister the %s task: %v : %s", constants.DaemonTaskName, err, stderr)
			return err
		}
	}
	return nil
}

func checkIfDaemonTaskRunning() error {
	stdout, stderr, err := powershell.Execute(fmt.Sprintf(`(Get-ScheduledTask -TaskName "%s").State`, constants.DaemonTaskName))
	if err != nil {
		logging.Debugf("%s task is not running: %v : %s", constants.DaemonTaskName, err, stderr)
		return err
	}
	if strings.TrimSpace(stdout) != "Running" {
		return fmt.Errorf("expected %s task to be in 'Running' but got '%s'", constants.DaemonTaskName, stdout)
	}
	return nil
}

func fixDaemonTaskRunning() error {
	if daemonRunning() {
		if err := killDaemonProcess(); err != nil {
			return err
		}
	}
	_, stderr, err := powershell.Execute("Start-ScheduledTask", "-TaskName", constants.DaemonTaskName)
	if err != nil {
		logging.Debugf("unable to run the %s task: %v : %s", constants.DaemonTaskName, err, stderr)
		return err
	}
	return waitForDaemonRunning()
}

func checkIfOlderTask() error {
	stdout, stderr, err := powershell.Execute(fmt.Sprintf(`(Get-ScheduledTask -TaskName "%s").Version`, constants.DaemonTaskName))
	if err != nil {
		return fmt.Errorf("%s task is not running: %v : %s", constants.DaemonTaskName, err, stderr)
	}
	if strings.TrimSpace(stdout) != version.GetCRCVersion() {
		return fmt.Errorf("%w but got '%s'", errOlderVersion, stdout)
	}
	return nil
}

var daemonPoshScriptPath = filepath.Join(constants.CrcBinDir, "hidden_daemon.ps1")

func getDaemonPoshScriptContent() []byte {
	binPath, err := os.Executable()
	if err != nil {
		return []byte{}
	}
	daemonCmdArgs := `daemon --log-level debug`
	return []byte(fmt.Sprintf(daemonPoshScriptTemplate, binPath, daemonCmdArgs))
}

func checkDaemonPoshScript() error {
	if exists := crcos.FileExists(daemonPoshScriptPath); exists {
		// check the script contains the path to the current executable
		if err := crcos.FileContentMatches(daemonPoshScriptPath, getDaemonPoshScriptContent()); err == nil {
			return nil
		}
	}
	return fmt.Errorf("Powershell script for running the daemon does not exist")
}

func fixDaemonPoshScript() error {
	return os.WriteFile(daemonPoshScriptPath, getDaemonPoshScriptContent(), 0600)
}
