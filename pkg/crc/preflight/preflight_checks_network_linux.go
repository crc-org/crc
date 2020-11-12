// +build linux

package preflight

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

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
		fixDescription:   "Checking if NetworkManager is installed",
		fix:              fixNetworkManagerInstalled,
	},
	{
		configKeySuffix:  "check-network-manager-running",
		checkDescription: "Checking if NetworkManager service is running",
		check:            checkNetworkManagerIsRunning,
		fixDescription:   "Checking if NetworkManager service is running",
		fix:              fixNetworkManagerIsRunning,
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

func checkCrcDnsmasqConfigFile() error {
	logging.Debug("Checking dnsmasq configuration")
	c := []byte(crcDnsmasqConfig)
	_, err := os.Stat(crcDnsmasqConfigPath)
	if err != nil {
		return fmt.Errorf("File not found: %s: %s", crcDnsmasqConfigPath, err.Error())
	}
	config, err := ioutil.ReadFile(filepath.Clean(crcDnsmasqConfigPath))
	if err != nil {
		return fmt.Errorf("Error opening file: %s: %s", crcDnsmasqConfigPath, err.Error())
	}
	if !bytes.Equal(config, c) {
		return fmt.Errorf("Config file contains changes: %s", crcDnsmasqConfigPath)
	}
	logging.Debug("dnsmasq configuration is good")
	return nil
}

func fixCrcDnsmasqConfigFile() error {
	logging.Debug("Fixing dnsmasq configuration")
	err := crcos.WriteToFileAsRoot(
		fmt.Sprintf("write dnsmasq configuration in %s", crcDnsmasqConfigPath),
		crcDnsmasqConfig,
		crcDnsmasqConfigPath,
		0644,
	)
	if err != nil {
		return fmt.Errorf("Failed to write dnsmasq config file: %s: %v", crcDnsmasqConfigPath, err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if err := sd.Reload("NetworkManager"); err != nil {
		return fmt.Errorf("Failed to restart NetworkManager: %v", err)
	}

	logging.Debug("dnsmasq configuration fixed")
	return nil
}

func removeCrcDnsmasqConfigFile() error {
	if err := checkNetworkManagerInstalled(); err != nil {
		// When NetworkManager is not installed, this file won't exist
		return nil
	}
	// Delete the `crcDnsmasqConfigPath` file if exists,
	// ignore all the os PathError except `IsNotExist` one.
	if _, err := os.Stat(crcDnsmasqConfigPath); !os.IsNotExist(err) {
		logging.Debug("Removing dnsmasq configuration")
		err := crcos.RemoveFileAsRoot(
			fmt.Sprintf("removing dnsmasq configuration in %s", crcDnsmasqConfigPath),
			crcDnsmasqConfigPath,
		)
		if err != nil {
			return fmt.Errorf("Failed to remove dnsmasq config file: %s: %v", crcDnsmasqConfigPath, err)
		}

		logging.Debug("Reloading NetworkManager")
		sd := systemd.NewHostSystemdCommander()
		if err := sd.Reload("NetworkManager"); err != nil {
			return fmt.Errorf("Failed to restart NetworkManager: %v", err)
		}
	}
	return nil
}

func checkCrcNetworkManagerConfig() error {
	logging.Debug("Checking NetworkManager configuration")
	c := []byte(crcNetworkManagerConfig)
	_, err := os.Stat(crcNetworkManagerConfigPath)
	if err != nil {
		return fmt.Errorf("File not found: %s: %s", crcNetworkManagerConfigPath, err.Error())
	}
	config, err := ioutil.ReadFile(filepath.Clean(crcNetworkManagerConfigPath))
	if err != nil {
		return fmt.Errorf("Error opening file: %s: %s", crcNetworkManagerConfigPath, err.Error())
	}
	if !bytes.Equal(config, c) {
		return fmt.Errorf("Config file contains changes: %s", crcNetworkManagerConfigPath)
	}
	logging.Debug("NetworkManager configuration is good")
	return nil
}

func fixCrcNetworkManagerConfig() error {
	logging.Debug("Fixing NetworkManager configuration")
	err := crcos.WriteToFileAsRoot(
		fmt.Sprintf("write NetworkManager config in %s", crcNetworkManagerConfigPath),
		crcNetworkManagerConfig,
		crcNetworkManagerConfigPath,
		0644,
	)
	if err != nil {
		return fmt.Errorf("Failed to write NetworkManager config file: %s: %v", crcNetworkManagerConfigPath, err)
	}

	logging.Debug("Reloading NetworkManager")
	sd := systemd.NewHostSystemdCommander()
	if err := sd.Reload("NetworkManager"); err != nil {
		return fmt.Errorf("Failed to restart NetworkManager: %v", err)
	}

	logging.Debug("NetworkManager configuration fixed")
	return nil
}

func removeCrcNetworkManagerConfig() error {
	if err := checkNetworkManagerInstalled(); err != nil {
		// When NetworkManager is not installed, this file won't exist
		return nil
	}
	if _, err := os.Stat(crcNetworkManagerConfigPath); !os.IsNotExist(err) {
		logging.Debug("Removing NetworkManager configuration")
		err := crcos.RemoveFileAsRoot(
			fmt.Sprintf("Removing NetworkManager config in %s", crcNetworkManagerConfigPath),
			crcNetworkManagerConfigPath,
		)
		if err != nil {
			return fmt.Errorf("Failed to remove NetworkManager config file: %s: %v", crcNetworkManagerConfigPath, err)
		}

		logging.Debug("Reloading NetworkManager")
		sd := systemd.NewHostSystemdCommander()
		if err := sd.Reload("NetworkManager"); err != nil {
			return fmt.Errorf("Failed to restart NetworkManager: %v", err)
		}
	}
	return nil
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

func fixNetworkManagerInstalled() error {
	return fmt.Errorf("NetworkManager is required and must be installed manually")
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

func fixNetworkManagerIsRunning() error {
	return fmt.Errorf("NetworkManager is required. Please make sure it is installed and running manually")
}
