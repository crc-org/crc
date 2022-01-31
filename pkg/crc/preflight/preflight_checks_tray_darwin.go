package preflight

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os/launchd"
)

const (
	daemonAgentLabel = "crc.daemon"
	trayAgentLabel   = "crc.tray"
)

type TrayVersion struct {
	ShortVersion string `plist:"CFBundleShortVersionString"`
}

func getTrayConfig() (*launchd.AgentConfig, error) {
	stdOutFilePathTray := filepath.Join(constants.CrcBaseDir, ".crct-agent.log")

	trayConfig := launchd.AgentConfig{
		Label:          trayAgentLabel,
		ExecutablePath: constants.TrayExecutablePath(),
		StdOutFilePath: stdOutFilePathTray,
	}

	return &trayConfig, nil
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
	if launchd.PlistExists(daemonAgentLabel) {
		return errors.New("crc daemon plist should not exist anymore")
	}
	if launchd.AgentRunning(daemonAgentLabel) {
		return fmt.Errorf("crc daemon should not run anymore")
	}
	return nil
}

func unLoadDaemonAgent() error {
	_ = launchd.UnloadPlist(daemonAgentLabel)
	_ = launchd.RemovePlist(daemonAgentLabel)
	return nil
}

func checkIfTrayAgentRunning() error {
	if version.IsInstaller() {
		return nil
	}
	if !launchd.AgentRunning(trayAgentLabel) {
		return fmt.Errorf("Tray is not running")
	}
	return checkIfDaemonRunning()
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

func fixPlistFileExists(agentConfig launchd.AgentConfig) error {
	logging.Debugf("Creating plist for %s", agentConfig.Label)
	err := launchd.CreatePlist(agentConfig)
	if err != nil {
		return err
	}
	// load plist
	if version.IsInstaller() {
		return nil
	}
	if err := launchd.LoadPlist(agentConfig.Label); err != nil {
		logging.Debug("failed while creating plist:", err.Error())
		return err
	}
	return nil
}

func checkIfDaemonRunning() error {
	return crcerrors.Retry(context.Background(), time.Second*3, func() error {
		if _, err := daemonclient.New().APIClient.Version(); err != nil {
			return &crcerrors.RetriableError{Err: fmt.Errorf("Daemon is not yet running: %w", err)}
		}
		return nil
	}, time.Second)
}
