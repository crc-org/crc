// +build linux

package preflight

import (
	"fmt"
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
		configKeySuffix:  "check-systemd-networkd-running",
		checkDescription: "Checking if systemd-networkd is running",
		check:            checkSystemdNetworkdIsNotRunning,
		fixDescription:   "Network configuration with systemd-networkd is not supported. Perhaps you can try this new network mode: https://github.com/code-ready/crc/wiki/VPN-support--with-an--userland-network-stack",
		flags:            NoFix,
	},
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

var (
	crcNetworkManagerRootPath = filepath.Join(string(filepath.Separator), "etc", "NetworkManager")

	crcDnsmasqConfigPath = filepath.Join(crcNetworkManagerRootPath, "dnsmasq.d", "crc.conf")
	crcDnsmasqConfig     = `server=/apps-crc.testing/192.168.130.11
server=/crc.testing/192.168.130.11
`

	crcNetworkManagerConfigPath = filepath.Join(crcNetworkManagerRootPath, "conf.d", "crc-nm-dnsmasq.conf")
	crcNetworkManagerConfig     = `[main]
dns=dnsmasq
`

	crcNetworkManagerOldDispatcherPath = filepath.Join(crcNetworkManagerRootPath, "dispatcher.d", "pre-up.d", "99-crc.sh")
	crcNetworkManagerDispatcherPath    = filepath.Join(crcNetworkManagerRootPath, "dispatcher.d", "99-crc.sh")
	crcNetworkManagerDispatcherConfig  = `#!/bin/sh
# This is a NetworkManager dispatcher script to configure split DNS for
# the 'crc' libvirt network.
#
# The corresponding crc bridge is not created through NetworkManager, so
# it cannot be configured permanently through NetworkManager. We make the
# change directly using systemd-resolve instead.
#
# systemd-resolve is used instead of resolvectl due to distributions shipping
# systemd releases older than 239 not having the newer renamed tool. resolvectl
# supports being called as systemd-resolve, correctly handling the old CLI.
#
# NetworkManager will overwrite this systemd-resolve configuration every time a
# network connection goes up/down, so we run this script on each of these events
# to restore our settings. This is a NetworkManager bug which is fixed in
# version 1.26.6 by this commit:
# https://cgit.freedesktop.org/NetworkManager/NetworkManager/commit/?id=ee4e679bc7479de42780ebd8e3a4d74afa2b2ebe

export LC_ALL=C

systemd-resolve --interface crc --set-dns 192.168.130.11 --set-domain ~testing

exit 0
`
)

var systemdResolvedPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-dnsmasq-network-manager-config",
		checkDescription: "Checking if dnsmasq configurations file exist for NetworkManager",
		check:            checkCrcDnsmasqAndNetworkManagerConfigFile,
		fixDescription:   "Removing dnsmasq configuration file for NetworkManager",
		fix:              fixCrcDnsmasqAndNetworkManagerConfigFile,
	},
	{
		configKeySuffix:  "check-systemd-resolved-running",
		checkDescription: "Checking if the systemd-resolved service is running",
		check:            checkSystemdResolvedIsRunning,
		fixDescription:   "systemd-resolved is required on this distribution. Please make sure it is installed and running manually",
		flags:            NoFix,
	},
	{
		configKeySuffix:    "check-network-manager-dispatcher-file",
		checkDescription:   fmt.Sprintf("Checking if %s exists", crcNetworkManagerDispatcherPath),
		check:              checkCrcNetworkManagerDispatcherFile,
		fixDescription:     "Writing NetworkManager dispatcher file for crc",
		fix:                fixCrcNetworkManagerDispatcherFile,
		cleanupDescription: fmt.Sprintf("Removing %s file", crcNetworkManagerDispatcherPath),
		cleanup:            removeCrcNetworkManagerDispatcherFile,
	},
}

func fixNetworkManagerConfigFile(path string, content string, perms os.FileMode) error {
	err := crcos.WriteToFileAsRoot(
		fmt.Sprintf("Writing NetworkManager configuration to %s", path),
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
			fmt.Sprintf("Removing NetworkManager configuration file in %s", path),
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

func checkSystemdNetworkdIsNotRunning() error {
	err := checkSystemdServiceRunning("systemd-networkd.service")
	if err == nil {
		return fmt.Errorf("systemd-networkd.service is running")
	}

	logging.Debugf("systemd-networkd.service is not running")
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

func checkSystemdServiceRunning(service string) error {
	logging.Debugf("Checking if %s is running", service)
	sd := systemd.NewHostSystemdCommander()
	status, err := sd.Status(service)
	if err != nil {
		return err
	}
	if status != states.Running {
		return fmt.Errorf("%s is not running", service)
	}
	logging.Debugf("%s is already running", service)
	return nil
}

func checkNetworkManagerIsRunning() error {
	return checkSystemdServiceRunning("NetworkManager.service")
}

func checkSystemdResolvedIsRunning() error {
	return checkSystemdServiceRunning("systemd-resolved.service")
}

func checkCrcNetworkManagerDispatcherFile() error {
	logging.Debug("Checking NetworkManager dispatcher file for crc network")
	err := crcos.FileContentMatches(crcNetworkManagerDispatcherPath, []byte(crcNetworkManagerDispatcherConfig))
	if err != nil {
		return err
	}
	logging.Debug("Dispatcher file has the expected content")
	return nil
}

func fixCrcNetworkManagerDispatcherFile() error {
	logging.Debug("Fixing NetworkManager dispatcher configuration")

	// Remove dispatcher script which was used in crc 1.20 - it's been moved to a new location
	_ = removeNetworkManagerConfigFile(crcNetworkManagerOldDispatcherPath)

	err := fixNetworkManagerConfigFile(crcNetworkManagerDispatcherPath, crcNetworkManagerDispatcherConfig, 0755)
	if err != nil {
		return err
	}

	logging.Debug("NetworkManager dispatcher configuration fixed")
	return nil
}

func removeCrcNetworkManagerDispatcherFile() error {
	// Remove dispatcher script which was used in crc 1.20 - it's been moved to a new location
	_ = removeNetworkManagerConfigFile(crcNetworkManagerOldDispatcherPath)

	return removeNetworkManagerConfigFile(crcNetworkManagerDispatcherPath)
}

func checkCrcDnsmasqAndNetworkManagerConfigFile() error {
	// IF check return nil, which means file
	if _, err := os.Stat(crcDnsmasqConfigPath); !os.IsNotExist(err) {
		return fmt.Errorf("%s file exists", crcDnsmasqConfigPath)
	}
	if _, err := os.Stat(crcNetworkManagerConfigPath); !os.IsNotExist(err) {
		return fmt.Errorf("%s file exists", crcNetworkManagerConfigPath)
	}
	return nil
}

func fixCrcDnsmasqAndNetworkManagerConfigFile() error {
	// In case user upgrades from f-32 to f-33 the dnsmasq config for NM still
	// exists and needs to be removed.
	if err := removeCrcNetworkManagerConfig(); err != nil {
		logging.Debugf("%s: not present.", crcNetworkManagerConfigPath)
	}
	if err := removeCrcDnsmasqConfigFile(); err != nil {
		logging.Debugf("%s: not present.", crcDnsmasqConfigPath)
	}
	return nil
}
