package preflight

import (
	"bytes"
	"fmt"
	"io/ioutil"
	goos "os"
	"path/filepath"

	"github.com/Masterminds/semver"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/launchd"
	"howett.net/plist"
)

const (
	daemonAgentLabel = "crc.daemon"
	trayAgentLabel   = "crc.tray"
)

type TrayVersion struct {
	ShortVersion string `plist:"CFBundleShortVersionString"`
}

func checkIfDaemonPlistFileExists() error {
	// we have no way of detecting if a running crc daemon instance uses
	// the same binary as the one running 'crc setup'
	// Safer for now to keep the previous behaviour of always
	// recreating the plist/restarting the daemon
	// When we have a way of checking the version of the running
	// daemon instance, we can recreate the plist less often.

	return fmt.Errorf("Ignoring this check and triggering re-creation of daemon plist")

	/*
		daemonConfig, err := getDaemonConfig()
		if err != nil {
			return err
		}
		return launchd.CheckPlist(*daemonConfig)
	*/
}

func getDaemonConfig() (*launchd.AgentConfig, error) {
	stdOutFilePathDaemon := filepath.Join(constants.CrcBaseDir, ".crcd-agent.log")

	currentExecutablePath, err := goos.Executable()
	if err != nil {
		return nil, err
	}
	daemonConfig := launchd.AgentConfig{
		Label:          daemonAgentLabel,
		ExecutablePath: currentExecutablePath,
		StdOutFilePath: stdOutFilePathDaemon,
		Args:           []string{"daemon"},
	}

	return &daemonConfig, nil
}

func getTrayConfig() (*launchd.AgentConfig, error) {
	stdOutFilePathTray := filepath.Join(constants.CrcBaseDir, ".crct-agent.log")

	trayConfig := launchd.AgentConfig{
		Label:          trayAgentLabel,
		ExecutablePath: constants.TrayExecutablePath,
		StdOutFilePath: stdOutFilePathTray,
	}

	return &trayConfig, nil
}

func fixDaemonPlistFileExists() error {
	// Try to remove the daemon agent from launchd
	// and recreate its plist
	_ = launchd.Remove(daemonAgentLabel)

	daemonConfig, err := getDaemonConfig()
	if err != nil {
		return nil
	}

	return fixPlistFileExists(*daemonConfig)
}

func removeDaemonPlistFile() error {
	return launchd.RemovePlist(daemonAgentLabel)
}

func checkIfTrayPlistFileExists() error {
	trayConfig, err := getTrayConfig()
	if err != nil {
		return err
	}
	return launchd.CheckPlist(*trayConfig)
}

func fixTrayPlistFileExists() error {
	trayConfig, err := getTrayConfig()
	if err != nil {
		return err
	}
	return fixPlistFileExists(*trayConfig)
}

func removeTrayPlistFile() error {
	return launchd.RemovePlist(trayAgentLabel)
}

func checkIfDaemonAgentRunning() error {
	if !launchd.AgentRunning(daemonAgentLabel) {
		return fmt.Errorf("crc daemon is not running")
	}
	return nil
}

func fixDaemonAgentRunning() error {
	logging.Debug("Starting daemon agent")
	if err := launchd.LoadPlist(daemonAgentLabel); err != nil {
		return err
	}
	return launchd.StartAgent(daemonAgentLabel)
}

func unLoadDaemonAgent() error {
	return launchd.UnloadPlist(daemonAgentLabel)
}

func checkIfTrayAgentRunning() error {
	if !launchd.AgentRunning(trayAgentLabel) {
		return fmt.Errorf("Tray is not running")
	}
	return nil
}

func fixTrayAgentRunning() error {
	logging.Debug("Starting tray agent")
	if err := launchd.LoadPlist(trayAgentLabel); err != nil {
		return err
	}
	return launchd.StartAgent(trayAgentLabel)
}

func unLoadTrayAgent() error {
	return launchd.UnloadPlist(trayAgentLabel)
}

func checkTrayVersion() error {
	v, err := getTrayVersion(constants.TrayAppBundlePath)
	if err != nil {
		logging.Error(err.Error())
		return err
	}
	currentVersion, err := semver.NewVersion(v)
	if err != nil {
		logging.Error(err.Error())
		return err
	}
	expectedVersion, err := semver.NewVersion(version.GetCRCMacTrayVersion())
	if err != nil {
		logging.Error(err.Error())
		return err
	}

	if expectedVersion.GreaterThan(currentVersion) {
		return fmt.Errorf("Cached version is older then latest version: %s < %s", currentVersion.String(), expectedVersion.String())
	}
	return nil
}

func fixTrayVersion() error {
	// get the tray app
	err := downloadOrExtractTrayApp(constants.GetCRCMacTrayDownloadURL(), constants.CrcBinDir)
	if err != nil {
		return err
	}
	return launchd.RestartAgent(trayAgentLabel)
}

func checkTrayExecutablePresent() error {
	if !os.FileExists(constants.TrayExecutablePath) {
		return fmt.Errorf("Tray executable does not exist")
	}
	return nil
}

func fixTrayExecutablePresent() error {
	return downloadOrExtractTrayApp(constants.GetCRCMacTrayDownloadURL(), constants.CrcBinDir)
}

func fixPlistFileExists(agentConfig launchd.AgentConfig) error {
	logging.Debugf("Creating plist for %s", agentConfig.Label)
	err := launchd.CreatePlist(agentConfig)
	if err != nil {
		return err
	}
	// load plist
	if err := launchd.LoadPlist(agentConfig.Label); err != nil {
		logging.Debug("failed while creating plist:", err.Error())
		return err
	}
	return nil
}

func getTrayVersion(trayAppPath string) (string, error) {
	var version TrayVersion
	f, err := ioutil.ReadFile(filepath.Join(trayAppPath, "Contents", "Info.plist")) // #nosec G304
	if err != nil {
		return "", err
	}
	decoder := plist.NewDecoder(bytes.NewReader(f))
	err = decoder.Decode(&version)
	if err != nil {
		return "", err
	}

	return version.ShortVersion, nil
}
