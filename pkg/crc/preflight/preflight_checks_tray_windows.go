package preflight

import (
	"fmt"
	goos "os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/input"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
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
	// try to remove if a previous version exists
	_ = service.Stop(constants.DaemonServiceName)
	_ = service.Delete(constants.DaemonServiceName)
	// get executables path
	binPath, err := goos.Executable()
	if err != nil {
		return fmt.Errorf("Unable to find the current executables location: %v", err)
	}
	binPathWithArgs := fmt.Sprintf("%s daemon --log-level debug", strings.TrimSpace(binPath))

	// get the account name
	whoami, err := exec.LookPath("whoami.exe")
	if err != nil {
		return fmt.Errorf("Unable to find whoami.exe in path: %v", err)
	}
	stdOut, err := exec.Command(whoami).Output() // #nosec
	if err != nil {
		return fmt.Errorf("Unable to get the current account name: %v", err)
	}
	accountName := string(stdOut)
	accountName = strings.TrimSpace(accountName)

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
	err := service.Stop(constants.DaemonServiceName)
	if err != nil {
		return fmt.Errorf("Failed to stop the daemon service: %v", err)
	}
	return service.Delete(constants.DaemonServiceName)
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
	tray := cache.NewWindowsTrayCache(constants.TrayBinaryDir)
	if err := tray.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to cache system tray: %v", err)
	}
	logging.Debugf("System tray is cached")
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
	// Kill older running tray
	if err := checkTrayRunning(); err == nil {
		if err := stopTray(); err != nil {
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
	return goos.Remove(filepath.Join(startUpFolder, constants.TrayShortcutName))
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
	err := exec.Command(constants.TrayBinaryPath).Start()
	if err != nil {
		return err
	}
	return nil
}

func stopTray() error {
	cmd := fmt.Sprintf("Stop-Process -Name %s", trayProcessName)
	_, _, err := powershell.Execute(cmd)
	if err != nil {
		return fmt.Errorf("Failed to stop running tray: %v", err)
	}
	return nil
}
