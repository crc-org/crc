package preflight

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/cache"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/version"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
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
      <Command>%s</Command>
      <Arguments>%s</Arguments>
    </Exec>
  </Actions>
</Task>
`
	errOlderVersion = fmt.Errorf("expected %s task to be on version '%s'", constants.DaemonTaskName, version.GetCRCVersion())
)

func genDaemonTaskInstallTemplate(crcVersion, userName, backgroundLauncherPath, daemonCommand string) (string, error) {
	var escapedDaemonCommand, escapedBackgroundLauncherPath bytes.Buffer
	if err := xml.EscapeText(&escapedDaemonCommand, []byte(daemonCommand)); err != nil {
		return "", err
	}
	if err := xml.EscapeText(&escapedBackgroundLauncherPath, []byte(backgroundLauncherPath)); err != nil {
		return "", err
	}

	return fmt.Sprintf(daemonTaskTemplate,
		crcVersion,
		userName,
		escapedBackgroundLauncherPath.String(),
		escapedDaemonCommand.String(),
	), nil
}

func checkIfDaemonTaskInstalled() error {
	_, stderr, err := powershell.Execute("Get-ScheduledTask", "-TaskName", constants.DaemonTaskName)
	if err != nil {
		logging.Debugf("%s task is not installed: %v : %s", constants.DaemonTaskName, err, stderr)
		return err
	}

	return checkIfOlderTask()
}

func fixDaemonTaskInstalled() error {
	// Remove older task if exist
	if err := removeDaemonTask(); err != nil {
		return err
	}
	crcBinPath, err := os.Executable()
	if err != nil {
		return err
	}

	if !crcos.FileExists(constants.Win32BackgroundLauncherPath()) {
		return fmt.Errorf("Missing background launcher binary at: %s", constants.Win32BackgroundLauncherPath())
	}

	binPathWithArgs := fmt.Sprintf(`"%s" daemon`, crcBinPath)
	// Get current user along with domain
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	taskContent, err := genDaemonTaskInstallTemplate(
		version.GetCRCVersion(),
		u.Username,
		constants.Win32BackgroundLauncherPath(),
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

func killDaemonProcessIfRunning() error {
	if daemonRunning() {
		if err := killDaemonProcess(); err != nil {
			return err
		}
	}
	return nil
}

func checkWin32BackgroundLauncherInstalled() error {
	c := cache.NewWin32BackgroundLauncherCache()
	return c.CheckVersion()
}

func fixWin32BackgroundLauncherInstalled() error {
	c := cache.NewWin32BackgroundLauncherCache()
	return c.EnsureIsCached()
}

func removeWin32BackgroundLauncher() error {
	if version.IsInstaller() {
		return nil
	}
	if crcos.FileExists(constants.Win32BackgroundLauncherPath()) {
		return os.Remove(constants.Win32BackgroundLauncherPath())
	}
	return nil
}
