package preflight

import (
	"fmt"
	"io/ioutil"
	goos "os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	dl "github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	"github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
	"github.com/code-ready/crc/pkg/os/windows/secpol"
	"github.com/code-ready/crc/pkg/os/windows/service"
)

var (
	startUpFolder   = filepath.Join(constants.GetHomeDir(), "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	trayProcessName = constants.TrayBinaryName[:len(constants.TrayBinaryName)-4]
)

func checkIfDaemonServiceInstalled() error {
	// We want to force installation whenever setup is ran
	// as we want the service config to point to the binary
	// with which the setup command was issued
	return fmt.Errorf("Ignoring check and forcing installation of daemon service")
}

func fixDaemonServiceInstalled() error {
	// only try to remove if a previous version exists
	if service.IsInstalled(constants.DaemonServiceName) {
		if service.IsRunning(constants.DaemonServiceName) {
			_ = service.Stop(constants.DaemonServiceName)
		}
		_ = service.Delete(constants.DaemonServiceName)
	}

	// get executables path
	binPath, err := goos.Executable()
	if err != nil {
		return fmt.Errorf("Unable to find the current executables location: %v", err)
	}
	binPathWithArgs := fmt.Sprintf("%s daemon", strings.TrimSpace(binPath))

	// get the account name
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("Failed to get username: %v", err)
	}
	logging.Debug("Got username: ", u.Username)
	accountName := strings.TrimSpace(u.Username)

	// get the password from user
	password, err := input.PromptUserForSecret("Enter account login password for service installation", "This is the login password of your current account, needed to install the daemon service")
	if err != nil {
		return fmt.Errorf("Unable to get login password: %v", err)
	}
	err = service.Create(constants.DaemonServiceName, binPathWithArgs, accountName, password)
	if err != nil {
		return fmt.Errorf("Failed to install CodeReady Containers daemon service: %v", err)
	}
	return nil
}

func removeDaemonService() error {
	// try to remove service if only it is installed
	// this should remove unnecessary UAC prompts during cleanup
	// service.IsInstalled doesn't need an admin shell to run
	if service.IsInstalled(constants.DaemonServiceName) {
		err := service.Stop(constants.DaemonServiceName)
		if err != nil {
			return fmt.Errorf("Failed to stop the daemon service: %v", err)
		}
		return service.Delete(constants.DaemonServiceName)
	}
	return nil
}

func checkIfDaemonServiceRunning() error {
	if service.IsRunning(constants.DaemonServiceName) {
		return nil
	}
	return fmt.Errorf("CodeReady Containers dameon service is not running")
}

func fixDaemonServiceRunning() error {
	return service.Start(constants.DaemonServiceName)
}

func checkTrayBinaryExists() error {
	if os.FileExists(constants.TrayBinaryPath) {
		return nil
	}
	return fmt.Errorf("Tray binary does not exists")
}

func fixTrayBinaryExists() error {
	tmpArchivePath, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Error("Failed creating temporary directory for extracting tray")
		return err
	}
	defer func() {
		_ = goos.RemoveAll(tmpArchivePath)
	}()

	logging.Debug("Trying to extract tray from crc binary")
	err = embed.Extract(filepath.Base(constants.GetCRCWindowsTrayDownloadURL()), tmpArchivePath)
	if err != nil {
		logging.Debug("Could not extract tray from crc binary", err)
		logging.Debug("Downloading crc tray")
		_, err = dl.Download(constants.GetCRCWindowsTrayDownloadURL(), tmpArchivePath, 0600)
		if err != nil {
			return err
		}
	}
	archivePath := filepath.Join(tmpArchivePath, filepath.Base(constants.GetCRCWindowsTrayDownloadURL()))
	_, err = extract.Uncompress(archivePath, constants.TrayBinaryDir)
	if err != nil {
		return fmt.Errorf("Cannot uncompress '%s': %v", archivePath, err)
	}

	return nil
}

func checkTrayBinaryVersion() error {
	versionCmd := `(Get-Item %s).VersionInfo.ProductVersion`
	stdOut, stdErr, err := powershell.Execute(fmt.Sprintf(versionCmd, constants.TrayBinaryPath))
	if err != nil {
		return fmt.Errorf("Failed to get the version of tray: %v: %s", err, stdErr)
	}
	currentTrayVersion := strings.TrimSpace(stdOut)
	if currentTrayVersion != version.GetCRCWindowsTrayVersion() {
		return fmt.Errorf("Current tray version doesn't match with expected version")
	}
	return nil
}

func fixTrayBinaryVersion() error {
	// If a tray is already running kill it
	if err := checkTrayRunning(); err == nil {
		cmd := `Stop-Process -Name "tray-windows"`
		if _, _, err := powershell.Execute(cmd); err != nil {
			logging.Debugf("Failed to kill running tray: %v", err)
		}
	}
	return fixTrayBinaryExists()
}

func checkTrayBinaryAddedToStartupFolder() error {
	if os.FileExists(filepath.Join(startUpFolder, constants.TrayShortcutName)) {
		return nil
	}
	return fmt.Errorf("Tray shortcut does not exists in startup folder")
}

func fixTrayBinaryAddedToStartupFolder() error {
	cmd := fmt.Sprintf(
		"New-Item -ItemType SymbolicLink -Path \"%s\" -Name \"%s\" -Value \"%s\"",
		startUpFolder,
		constants.TrayShortcutName,
		constants.TrayBinaryPath,
	)
	_, _, err := powershell.ExecuteAsAdmin("Adding tray binary to startup applications", cmd)
	if err != nil {
		return fmt.Errorf("Failed to create shortcut of tray binary in startup folder: %v", err)
	}
	return nil
}

func removeTrayBinaryFromStartupFolder() error {
	err := goos.Remove(filepath.Join(startUpFolder, constants.TrayShortcutName))
	if err != nil {
		logging.Warn("Failed to remove tray from startup folder: ", err)
	}
	return nil
}

func checkTrayRunning() error {
	cmd := fmt.Sprintf("Get-Process -Name \"%s\"", trayProcessName)
	_, stdErr, err := powershell.Execute(cmd)
	if err != nil {
		return fmt.Errorf("Failed to check if the tray is running: %v", err)
	}
	if strings.Contains(stdErr, constants.TrayBinaryName) {
		return fmt.Errorf("Tray binary is not running")
	}
	return nil
}

func fixTrayRunning() error {
	// #nosec G204
	err := exec.Command(constants.TrayBinaryPath).Start()
	if err != nil {
		return err
	}
	return nil
}

func stopTray() error {
	if err := checkTrayRunning(); err != nil {
		logging.Debug("Failed to check if a tray is running: ", err)
		return nil
	}
	cmd := fmt.Sprintf("Stop-Process -Name %s", trayProcessName)
	_, _, err := powershell.Execute(cmd)
	if err != nil {
		logging.Warn("Failed to stop running tray: ", err)
	}
	return nil
}

func checkUserHasServiceLogonEnabled() error {
	username := getCurrentUsername()
	enabled, err := secpol.UserAllowedToLogonAsService(username)
	if enabled {
		return nil
	}
	return fmt.Errorf("Failed to check if user is allowed to log on as service: %v", err)
}

func fixUserHasServiceLogonEnabled() error {
	// get the username
	username := getCurrentUsername()

	return secpol.AllowUserToLogonAsService(username)
}

func getCurrentUsername() string {
	user, err := user.Current()
	if err != nil {
		logging.Debugf("Unable to get current user's name: %v", err)
		return ""
	}
	return strings.TrimSpace(strings.Split(user.Username, "\\")[1])
}

func disableUserServiceLogon() error {
	username := getCurrentUsername()

	if err := secpol.RemoveLogonAsServiceUserRight(username); err != nil {
		logging.Warn("Failed trying to remove log on as a service user right: ", err)
	}
	return nil
}
