package preflight

import (
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
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
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckDefaultPool.Name),
		checkDefaultPoolAvailable,
		"Checking if default pool is available",
		config.GetBool(cmdConfig.WarnCheckDefaultPool.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckDefaultPoolSpace.Name),
		checkDefaultPoolHasSufficientSpace,
		"Checking if default pool has sufficient free space",
		config.GetBool(cmdConfig.WarnCheckDefaultPoolSpace.Name),
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
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	preflightCheckAndFix(checkVirtualizationEnabled,
		fixVirtualizationEnabled,
		"Setting up virtualization",
	)
	preflightCheckAndFix(checkKvmEnabled,
		fixKvmEnabled,
		"Setting up KVM",
	)
	preflightCheckAndFix(checkLibvirtInstalled,
		fixLibvirtInstalled,
		"Installing Libvirt",
	)
	preflightCheckAndFix(checkUserPartOfLibvirtGroup,
		fixUserPartOfLibvirtGroup,
		"Adding user to libvirt group",
	)
	preflightCheckAndFix(checkLibvirtEnabled,
		fixLibvirtEnabled,
		"Enabling libvirt",
	)
	preflightCheckAndFix(checkLibvirtServiceRunning,
		fixLibvirtServiceRunning,
		"Starting Libvirt service",
	)
	preflightCheckAndFix(checkMachineDriverLibvirtInstalled,
		fixMachineDriverLibvirtInstalled,
		"Installing crc-driver-libvirt",
	)
	preflightCheckAndFix(checkDefaultPoolAvailable,
		fixDefaultPoolAvailable,
		"Creating default storage pool",
	)
	preflightCheckAndFix(checkDefaultPoolHasSufficientSpace,
		fixDefaultPoolHasSufficientSpace,
		"Setting up default pool",
	)
	preflightCheckAndFix(checkLibvirtCrcNetworkAvailable,
		fixLibvirtCrcNetworkAvailable,
		"Setting up Libvirt crc network",
	)
	preflightCheckAndFix(checkLibvirtCrcNetworkActive,
		fixLibvirtCrcNetworkActive,
		"Starting Libvirt crc network",
	)
}
