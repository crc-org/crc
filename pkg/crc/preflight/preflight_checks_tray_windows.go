package preflight

import (
	"fmt"
	goos "os"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

var daemonBatchFileShortcutPath = filepath.Join(constants.StartupFolder, constants.DaemonBatchFileShortcutName)

func checkIfDaemonInstalled() error {
	if os.FileExists(constants.DaemonBatchFilePath) || os.FileExists(constants.DaemonPSScriptPath) {
		return fmt.Errorf("Daemon should not be installed")
	}
	if os.FileExists(daemonBatchFileShortcutPath) {
		return fmt.Errorf("Daemon should not be installed")
	}
	return nil
}

func stopTray() error {
	/* we changed the name of the tray executable to crc-tray from tray-windows
	 * this tries to stop the old tray process, can be removed after a  few releases
	 */
	_, _, _ = powershell.Execute(`Stop-Process -Name tray-windows`)
	if err := checkIfTrayRunning(); err != nil {
		return nil
	}
	trayProcessName := strings.TrimSuffix(constants.TrayExecutableName, ".exe")
	cmd := fmt.Sprintf(`Stop-Process -Name "%s"`, trayProcessName)
	if _, _, err := powershell.Execute(cmd); err != nil {
		return fmt.Errorf("Failed to kill tray process: %w", err)
	}
	return nil
}

func removeDaemon() error {
	_ = stopDaemon()
	var mErr errors.MultiError
	if err := os.RemoveFileIfExists(constants.DaemonBatchFilePath); err != nil {
		mErr.Collect(err)
	}
	if err := os.RemoveFileIfExists(constants.DaemonPSScriptPath); err != nil {
		mErr.Collect(err)
	}
	if err := os.RemoveFileIfExists(daemonBatchFileShortcutPath); err != nil {
		mErr.Collect(err)
	}
	if len(mErr.Errors) == 0 {
		return nil
	}
	return mErr
}

func stopDaemon() error {
	// get the PID of the process running command 'crc daemon --log-level debug'
	getCrcProcessCmd := `(Get-WmiObject Win32_Process -Filter "name = 'crc.exe'")`
	cmd := fmt.Sprintf(`(%s | Where-Object -Property CommandLine -Like '* daemon --log-level debug').ProcessId`,
		getCrcProcessCmd,
	)
	stdOut, _, err := powershell.Execute(cmd)
	if err != nil {
		return fmt.Errorf("Failed to get PID of the daemon process: %w", err)
	}
	logging.Debugf("PID of daemon: %s", strings.TrimSpace(stdOut))

	// kill the daemon process with the PID
	cmd = fmt.Sprintf(`Stop-Process -Id %s`, strings.TrimSpace(stdOut))
	if _, _, err := powershell.Execute(cmd); err != nil {
		return fmt.Errorf("Failed to kill the daemon process: %w", err)
	}
	return nil
}

func checkTrayExecutableExists() error {
	if os.FileExists(constants.TrayExecutablePath) && checkTrayVersion() {
		return nil
	}
	return fmt.Errorf("Tray executable does not exists")
}

func fixTrayExecutableExists() error {
	/* we changed the name of the tray executable to crc-tray.exe from tray-windows.exe
	 * this tries to remove the old tray folder, can be removed after a  few releases
	 */
	return goos.RemoveAll(filepath.Join(constants.CrcBinDir, "tray-windows"))
}

func checkTrayVersion() bool {
	cmd := fmt.Sprintf(`(Get-Item "%s").VersionInfo.FileVersion`, constants.TrayExecutablePath)
	stdOut, _, err := powershell.Execute(cmd)
	if err != nil {
		logging.Debugf("Failed to get the version of tray: %v", err)
		return false
	}
	logging.Debugf("Got tray version: %s", strings.TrimSpace(stdOut))
	return strings.TrimSpace(stdOut) == version.GetCRCWindowsTrayVersion()
}

func checkIfTrayRunning() error {
	cmd := fmt.Sprintf("Get-Process -Name %s",
		strings.TrimSuffix(constants.TrayExecutableName, filepath.Ext(constants.TrayExecutableName)))

	_, stdErr, err := powershell.Execute(cmd)
	if err != nil {
		return err
	}
	if strings.TrimSpace(stdErr) != "" {
		return fmt.Errorf("Tray not running: %s", stdErr)
	}
	return nil
}

func startTray() error {
	cmd := fmt.Sprintf(`Start-Process -FilePath "%s"`, constants.TrayExecutablePath)
	if _, _, err := powershell.Execute(cmd); err != nil {
		return fmt.Errorf("Failed to start tray process: %w", err)
	}
	return nil
}
