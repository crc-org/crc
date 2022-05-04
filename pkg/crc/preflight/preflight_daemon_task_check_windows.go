package preflight

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
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
  </Settings>
  <Triggers />
  <Actions Context="Author">
    <Exec>
      <Command>powershell.exe</Command>
      <Arguments>-WindowStyle Hidden -Command %s</Arguments>
    </Exec>
  </Actions>
</Task>
`
	errOlderVersion = fmt.Errorf("expected %s task to be on version '%s'", constants.DaemonTaskName, version.GetCRCVersion())
)

func genDaemonTaskInstallTemplate(crcVersion, daemonCommand string) (string, error) {
	var escapedName bytes.Buffer
	if err := xml.EscapeText(&escapedName, []byte(daemonCommand)); err != nil {
		return "", err
	}

	return fmt.Sprintf(daemonTaskTemplate,
		crcVersion,
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
	// prepare the task script
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to find the current executable location: %v", err)
	}
	binPathWithArgs := fmt.Sprintf("& '%s' daemon", binPath)

	taskContent, err := genDaemonTaskInstallTemplate(
		version.GetCRCVersion(),
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
	// First check if a daemon process is running and then check if
	// the running daemon process is of older version, this way  we
	// return nil if a daemon process is running that is not old
	if _, err := daemonclient.GetVersionFromDaemonAPI(); err == nil {
		if err := olderDaemonVersionRunning(); err != nil {
			return err
		}
		return nil
	}

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
	if err := olderDaemonVersionRunning(); err != nil {
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
