package preflight

import (
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	preflightCheckSucceedsOrFails(false,
		checkOcBinaryCached,
		"Checking if oc binary is cached",
		false,
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckVirtEnabled.Name),
		checkVirtualizationEnabled,
		"Checking if Virtualization is enabled",
		config.GetBool(cmdConfig.WarnCheckVirtEnabled.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckKvmEnabled.Name),
		checkKvmEnabled,
		"Checking if KVM is enabled",
		config.GetBool(cmdConfig.WarnCheckKvmEnabled.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckLibvirtInstalled.Name),
		checkLibvirtInstalled,
		"Checking if Libvirt is installed",
		config.GetBool(cmdConfig.WarnCheckLibvirtInstalled.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckUserInLibvirtGroup.Name),
		checkUserPartOfLibvirtGroup,
		"Checking if user is part of libvirt group",
		config.GetBool(cmdConfig.WarnCheckUserInLibvirtGroup.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckLibvirtEnabled.Name),
		checkLibvirtEnabled,
		"Checking if Libvirt is enabled",
		config.GetBool(cmdConfig.WarnCheckLibvirtEnabled.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckLibvirtRunning.Name),
		checkLibvirtServiceRunning,
		"Checking if Libvirt daemon is running",
		config.GetBool(cmdConfig.WarnCheckLibvirtRunning.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckLibvirtDriver.Name),
		checkMachineDriverLibvirtInstalled,
		"Checking if crc-driver-libvirt is installed",
		config.GetBool(cmdConfig.WarnCheckLibvirtDriver.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckCrcNetwork.Name),
		checkLibvirtCrcNetworkAvailable,
		"Checking if Libvirt crc network is available",
		config.GetBool(cmdConfig.WarnCheckCrcNetwork.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckCrcNetworkActive.Name),
		checkLibvirtCrcNetworkActive,
		"Checking if Libvirt crc network is active",
		config.GetBool(cmdConfig.WarnCheckCrcNetworkActive.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckCrcNetworkManagerConfig.Name),
		checkCrcNetworkManagerConfig,
		"Checking if /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf exists",
		config.GetBool(cmdConfig.WarnCheckCrcDnsmasqFile.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckCrcDnsmasqFile.Name),
		checkCrcDnsmasqConfigFile,
		"Checking if /etc/NetworkManager/dnsmasq.d/crc.conf exists",
		config.GetBool(cmdConfig.WarnCheckCrcDnsmasqFile.Name),
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	preflightCheckAndFix(false,
		checkOcBinaryCached,
		fixOcBinaryCached,
		"Caching oc binary",
		false,
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckVirtEnabled.Name),
		checkVirtualizationEnabled,
		fixVirtualizationEnabled,
		"Setting up virtualization",
		config.GetBool(cmdConfig.WarnCheckVirtEnabled.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckKvmEnabled.Name),
		checkKvmEnabled,
		fixKvmEnabled,
		"Setting up KVM",
		config.GetBool(cmdConfig.WarnCheckKvmEnabled.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckLibvirtInstalled.Name),
		checkLibvirtInstalled,
		fixLibvirtInstalled,
		"Installing Libvirt",
		config.GetBool(cmdConfig.WarnCheckLibvirtInstalled.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckUserInLibvirtGroup.Name),
		checkUserPartOfLibvirtGroup,
		fixUserPartOfLibvirtGroup,
		"Adding user to libvirt group",
		config.GetBool(cmdConfig.WarnCheckUserInLibvirtGroup.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckLibvirtEnabled.Name),
		checkLibvirtEnabled,
		fixLibvirtEnabled,
		"Enabling libvirt",
		config.GetBool(cmdConfig.WarnCheckLibvirtEnabled.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckLibvirtRunning.Name),
		checkLibvirtServiceRunning,
		fixLibvirtServiceRunning,
		"Starting Libvirt service",
		config.GetBool(cmdConfig.WarnCheckLibvirtRunning.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckLibvirtDriver.Name),
		checkMachineDriverLibvirtInstalled,
		fixMachineDriverLibvirtInstalled,
		"Installing crc-driver-libvirt",
		config.GetBool(cmdConfig.WarnCheckLibvirtDriver.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckCrcNetwork.Name),
		checkLibvirtCrcNetworkAvailable,
		fixLibvirtCrcNetworkAvailable,
		"Setting up Libvirt crc network",
		config.GetBool(cmdConfig.WarnCheckCrcNetwork.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckCrcNetworkActive.Name),
		checkLibvirtCrcNetworkActive,
		fixLibvirtCrcNetworkActive,
		"Starting Libvirt crc network",
		config.GetBool(cmdConfig.WarnCheckCrcNetworkActive.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckCrcNetworkManagerConfig.Name),
		checkCrcNetworkManagerConfig,
		fixCrcNetworkManagerConfig,
		"Writing Network Manager config for crc",
		config.GetBool(cmdConfig.WarnCheckCrcDnsmasqFile.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckCrcDnsmasqFile.Name),
		checkCrcDnsmasqConfigFile,
		fixCrcDnsmasqConfigFile,
		"Writing dnsmasq config for crc",
		config.GetBool(cmdConfig.WarnCheckCrcDnsmasqFile.Name),
	)
}
