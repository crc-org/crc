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
	dl "github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	"github.com/code-ready/crc/pkg/launchd"
	"github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
	"howett.net/plist"
)

const (
	daemonAgentLabel = "crc.daemon"
	trayAgentLabel   = "crc.tray"
)

var (
	stdOutFilePathDaemon = filepath.Join(constants.CrcBaseDir, ".crcd-agent.log")
	stdOutFilePathTray   = filepath.Join(constants.CrcBaseDir, ".crct-agent.log")
)

type TrayVersion struct {
	ShortVersion string `plist:"CFBundleShortVersionString"`
}

func checkIfDaemonPlistFileExists() error {
	if !launchd.PlistExists(daemonAgentLabel) {
		return fmt.Errorf("Daemon plist file does not exist")
	}
	return nil
}

func fixDaemonPlistFileExists() error {
	currentExecutablePath, err := goos.Executable()
	if err != nil {
		return err
	}
	daemonConfig := launchd.AgentConfig{
		Label:          daemonAgentLabel,
		BinaryPath:     currentExecutablePath,
		StdOutFilePath: stdOutFilePathDaemon,
		Args:           []string{"daemon", "--log-level", "debug"},
	}
	return fixPlistFileExists(daemonConfig)
}

func checkIfTrayPlistFileExists() error {
	if !launchd.PlistExists(trayAgentLabel) {
		return fmt.Errorf("Tray plist file does not exist")
	}
	return nil
}

func fixTrayPlistFileExists() error {
	trayConfig := launchd.AgentConfig{
		Label:          trayAgentLabel,
		BinaryPath:     constants.TrayBinaryPath,
		StdOutFilePath: stdOutFilePathTray,
	}
	return fixPlistFileExists(trayConfig)
}

func checkIfDaemonAgentRunning() error {
	if !launchd.AgentRunning(daemonAgentLabel) {
		return fmt.Errorf("crc daemon is not running")
	}
	return nil
}

func fixDaemonAgentRunning() error {
	return launchd.StartAgent(daemonAgentLabel)
}

func checkIfTrayAgentRunning() error {
	if !launchd.AgentRunning(trayAgentLabel) {
		return fmt.Errorf("Tray is not running")
	}
	return nil
}

func fixTrayAgentRunning() error {
	return launchd.StartAgent(trayAgentLabel)
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
	expectedVersion, err := semver.NewVersion(version.GetCRCTrayVersion())
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
	err := downloadOrExtractTrayApp()
	if err != nil {
		return err
	}
	return launchd.RestartAgent(trayAgentLabel)
}

func checkTrayBinaryPresent() error {
	if !os.FileExists(constants.TrayBinaryPath) {
		return fmt.Errorf("Tray binary does not exist")
	}
	return nil
}

func fixTrayBinaryPresent() error {
	return downloadOrExtractTrayApp()
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

func downloadOrExtractTrayApp() error {
	// Extract the tray and put it in the bin directory.
	tmpArchivePath, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Error("Failed creating temporary directory for extracting tray")
		return err
	}
	defer func() {
		_ = goos.RemoveAll(tmpArchivePath)
	}()

	logging.Debug("Trying to extract tray from crc binary")
	err = embed.Extract(filepath.Base(constants.GetCrcTrayDownloadURL()), tmpArchivePath)
	if err != nil {
		logging.Debug("Could not extract tray from crc binary", err)
		logging.Debug("Downloading crc tray")
		_, err = dl.Download(constants.GetCrcTrayDownloadURL(), tmpArchivePath, 0600)
		if err != nil {
			return err
		}
	}
	archivePath := filepath.Join(tmpArchivePath, filepath.Base(constants.GetCrcTrayDownloadURL()))
	outputPath := constants.CrcBinDir
	err = goos.MkdirAll(outputPath, 0750)
	if err != nil && !goos.IsExist(err) {
		return errors.Wrap(err, "Cannot create the target directory.")
	}
	err = extract.Uncompress(archivePath, outputPath)
	if err != nil {
		return errors.Wrapf(err, "Cannot uncompress '%s'", archivePath)
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
