package preflight

import (
	"fmt"
	"io/ioutil"
	goos "os"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	dl "github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	"github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func checkIfTrayInstalled() error {
	if os.FileExists(filepath.Join(constants.StartupFolder, constants.TrayShortcutName)) {
		return nil
	}
	return fmt.Errorf("CodeReady Containers tray is not Installed")
}

func checkIfDaemonInstalled() error {
	// We want to force the installation of daemon as we want to
	// run the daemon from the crc executable which ran the setup
	return fmt.Errorf("Ignoring check, forcing installation of daemon")
}

func fixDaemonInstalled() error {
	// call stop daemon
	_ = stopDaemon()
	currentExecutablePath, err := goos.Executable()
	if err != nil {
		return fmt.Errorf("Failed to find current executables path: %w", err)
	}

	// Create the PowerShell script to launch daemon in the background with hidden window
	daemonLaunchPSScriptContent := fmt.Sprintf(`$psi = New-Object System.Diagnostics.ProcessStartInfo
$newproc = New-Object System.Diagnostics.Process
$psi.FileName = "%s"
$psi.Arguments = "%s"
$psi.CreateNoWindow = $true
$psi.WindowStyle = 'Hidden'
$newproc.StartInfo = $psi
$newproc.Start()
$newproc
`, currentExecutablePath, "daemon --log-level debug")

	err = ioutil.WriteFile(constants.DaemonPSScriptPath, []byte(daemonLaunchPSScriptContent), 0600)
	if err != nil {
		return fmt.Errorf("Failed to create required PowerShell script for the daemon: %w", err)
	}

	// Create the cmd file to start the PowerShell script at login
	daemonBatchFileContent := fmt.Sprintf(`%s -file "%s"`, powershell.LocatePowerShell(), constants.DaemonPSScriptPath)
	err = ioutil.WriteFile(constants.DaemonBatchFilePath, []byte(daemonBatchFileContent), 0600)
	if err != nil {
		return fmt.Errorf("Failed to create required batch file for the daemon: %w", err)
	}

	// Crete symlink to cmd file in the startup folder
	cmd := fmt.Sprintf(`New-Item -ItemType SymbolicLink -Path "%s" -Name "%s" -Value "%s"`,
		constants.StartupFolder,
		constants.DaemonBatchFileName,
		constants.DaemonBatchFilePath,
	)
	if _, _, err := powershell.ExecuteAsAdmin("Create symlink to daemon batch file in start-up folder", cmd); err != nil {
		return fmt.Errorf("Error trying to create symlink to tray in start-up folder: %w", err)
	}

	// Start the daemon process from the PowerShell script
	if _, _, err := powershell.Execute(constants.DaemonPSScriptPath); err != nil {
		return fmt.Errorf("Failed to start the daemon process: %w", err)
	}

	return nil
}

func fixTrayInstalled() error {
	/* Start the tray process and copy the executable to start-up folder
	 * "$USERPROFILE\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup"
	 */
	_ = stopTray()

	cmd := fmt.Sprintf(`New-Item -ItemType SymbolicLink -Path "%s" -Name "%s" -Value "%s"`,
		constants.StartupFolder,
		constants.TrayShortcutName,
		constants.TrayExecutablePath,
	)
	if _, _, err := powershell.ExecuteAsAdmin("Create symlink to tray in start-up folder", cmd); err != nil {
		return fmt.Errorf("Error trying to create symlink to tray in start-up folder: %w", err)
	}
	cmd = fmt.Sprintf(`Start-Process -FilePath "%s"`, constants.TrayExecutablePath)
	if _, _, err := powershell.Execute(cmd); err != nil {
		return fmt.Errorf("Failed to start tray process: %w", err)
	}
	return nil
}

func removeTray() error {
	_ = stopTray()
	return goos.Remove(filepath.Join(constants.StartupFolder, constants.TrayShortcutName))
}

func stopTray() error {
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
	if err := goos.Remove(constants.DaemonBatchFilePath); err != nil {
		mErr.Collect(err)
	}
	if err := goos.Remove(constants.DaemonPSScriptPath); err != nil {
		mErr.Collect(err)
	}
	if len(mErr.Errors) == 0 {
		return nil
	}
	return mErr
}

func stopDaemon() error {
	executablePath, err := goos.Executable()
	if err != nil {
		return fmt.Errorf("Failed to find current executable's path: %w", err)
	}
	// get the PID of the process running command 'crc daemon --log-level debug'
	getCrcProcessCmd := `(Get-WmiObject Win32_Process -Filter "name = 'crc.exe'")`
	cmd := fmt.Sprintf(`(%s | Where-Object -Property CommandLine -Eq '"%s" daemon --log-level debug').ProcessId`,
		getCrcProcessCmd,
		executablePath,
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
	if os.FileExists(constants.TrayExecutablePath) {
		return nil
	}
	return fmt.Errorf("Tray executable does not exists")
}

func fixTrayExecutableExists() error {
	tmpArchivePath, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Error("Failed creating temporary directory for extracting tray")
		return err
	}
	defer func() {
		_ = goos.RemoveAll(tmpArchivePath)
	}()

	logging.Debug("Trying to extract tray from crc executable")
	trayFileName := filepath.Base(constants.GetCRCWindowsTrayDownloadURL())
	trayDestFileName := filepath.Join(tmpArchivePath, trayFileName)
	err = embed.Extract(trayFileName, trayDestFileName)
	if err != nil {
		logging.Debug("Could not extract tray from crc executable", err)
		logging.Debug("Downloading crc tray")
		_, err = dl.Download(constants.GetCRCWindowsTrayDownloadURL(), tmpArchivePath, 0600)
		if err != nil {
			return err
		}
	}
	_, err = extract.Uncompress(trayDestFileName, constants.TrayExecutableDir, false)
	if err != nil {
		return fmt.Errorf("Cannot uncompress '%s': %v", trayDestFileName, err)
	}

	return nil
}
