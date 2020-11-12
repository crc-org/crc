// +build linux

package preflight

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
	crcos "github.com/code-ready/crc/pkg/os"
)

var nmPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-network-manager-installed",
		checkDescription: "Checking if NetworkManager is installed",
		check:            checkNetworkManagerInstalled,
		fixDescription:   "NetworkManager is required and must be installed manually",
		flags:            NoFix,
	},
	{
		configKeySuffix:  "check-network-manager-running",
		checkDescription: "Checking if NetworkManager service is running",
		check:            checkNetworkManagerIsRunning,
		fixDescription:   "NetworkManager is required. Please make sure it is installed and running manually",
		flags:            NoFix,
	},
}

var dnsmasqPreflightChecks = [...]Check{
	{
		configKeySuffix:    "check-network-manager-config",
		checkDescription:   "Checking if /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf exists",
		check:              checkCrcNetworkManagerConfig,
		fixDescription:     "Writing Network Manager config for crc",
		fix:                fixCrcNetworkManagerConfig,
		cleanupDescription: "Removing /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf file",
		cleanup:            removeCrcNetworkManagerConfig,
	},
	{
		configKeySuffix:    "check-crc-dnsmasq-file",
		checkDescription:   "Checking if /etc/NetworkManager/dnsmasq.d/crc.conf exists",
		check:              checkCrcDnsmasqConfigFile,
		fixDescription:     "Writing dnsmasq config for crc",
		fix:                fixCrcDnsmasqConfigFile,
		cleanupDescription: "Removing /etc/NetworkManager/dnsmasq.d/crc.conf file",
		cleanup:            removeCrcDnsmasqConfigFile,
	},
}

func fixNetworkManagerConfigFile(path string, content string, perms os.FileMode) error {
	err := crcos.WriteToFileAsRoot(
		fmt.Sprintf("write NetworkManager configuration to %s", path),
		content,
		path,
		perms,
	)
	if err != nil {
		return fmt.Errorf("Failed to write config file: %s: %v", path, err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if err := sd.Reload("NetworkManager"); err != nil {
		return fmt.Errorf("Failed to restart NetworkManager: %v", err)
	}

	return nil
}

func removeNetworkManagerConfigFile(path string) error {
	if err := checkNetworkManagerInstalled(); err != nil {
		// When NetworkManager is not installed, its config files won't exist
		return nil
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		logging.Debugf("Removing NetworkManager configuration file: %s", path)
		err := crcos.RemoveFileAsRoot(
			fmt.Sprintf("removing NetworkManager configuration file in %s", path),
			path,
		)
		if err != nil {
			return fmt.Errorf("Failed to remove NetworkManager configuration file: %s: %v", path, err)
		}

		logging.Debug("Reloading NetworkManager")
		sd := systemd.NewHostSystemdCommander()
		if err := sd.Reload("NetworkManager"); err != nil {
			return fmt.Errorf("Failed to restart NetworkManager: %v", err)
		}
	}
	return nil
}

func checkCrcDnsmasqConfigFile() error {
	logging.Debug("Checking dnsmasq configuration")
	err := crcos.FileContentMatches(crcDnsmasqConfigPath, []byte(crcDnsmasqConfig))
	if err != nil {
		return err
	}
	logging.Debug("dnsmasq configuration is good")
	return nil
}

func fixCrcDnsmasqConfigFile() error {
	logging.Debug("Fixing dnsmasq configuration")
	err := fixNetworkManagerConfigFile(crcDnsmasqConfigPath, crcDnsmasqConfig, 0644)
	if err != nil {
		return err
	}

	logging.Debug("dnsmasq configuration fixed")
	return nil
}

func removeCrcDnsmasqConfigFile() error {
	return removeNetworkManagerConfigFile(crcDnsmasqConfigPath)
}

func checkCrcNetworkManagerConfig() error {
	logging.Debug("Checking NetworkManager configuration")
	err := crcos.FileContentMatches(crcNetworkManagerConfigPath, []byte(crcNetworkManagerConfig))
	if err != nil {
		return err
	}
	logging.Debug("NetworkManager configuration is good")
	return nil
}

func fixCrcNetworkManagerConfig() error {
	logging.Debug("Fixing NetworkManager configuration")
	err := fixNetworkManagerConfigFile(crcNetworkManagerConfigPath, crcNetworkManagerConfig, 0644)
	if err != nil {
		return err
	}
	logging.Debug("NetworkManager configuration fixed")
	return nil
}

func removeCrcNetworkManagerConfig() error {
	return removeNetworkManagerConfigFile(crcNetworkManagerConfigPath)
}

func checkNetworkManagerInstalled() error {
	logging.Debug("Checking if 'nmcli' is available")
	path, err := exec.LookPath("nmcli")
	if err != nil {
		return fmt.Errorf("NetworkManager cli nmcli was not found in path")
	}
	logging.Debug("'nmcli' was found in ", path)
	return nil
}

func checkNetworkManagerIsRunning() error {
	logging.Debug("Checking if NetworkManager.service is running")
	sd := systemd.NewHostSystemdCommander()
	status, err := sd.Status("NetworkManager")
	if err != nil {
		return err
	}
	if status != states.Running {
		return fmt.Errorf("NetworkManager.service is not running")
	}
	logging.Debug("NetworkManager.service is already running")
	return nil
}
