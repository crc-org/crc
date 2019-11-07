package preflight

import (
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
)

var genericPreflightChecks = [...]PreflightCheck{
	{
		skipConfigName:   cmdConfig.SkipCheckRootUser.Name,
		warnConfigName:   cmdConfig.WarnCheckRootUser.Name,
		checkDescription: "Checking if running as non-root",
		check:            checkIfRunningAsNormalUser,
		fix:              fixRunAsNormalUser,
	},
	{
		checkDescription: "Checking if oc binary is cached",
		check:            checkOcBinaryCached,
		fixDescription:   "Caching oc binary",
		fix:              fixOcBinaryCached,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckBundleCached.Name,
		warnConfigName:   cmdConfig.WarnCheckBundleCached.Name,
		checkDescription: "Checking if CRC bundle is cached in '$HOME/.crc'",
		check:            checkBundleCached,
		fixDescription:   "Unpacking bundle from the CRC binary",
		fix:              fixBundleCached,
		flags:            SetupOnly,
	},
}

var libvirtPreflightChecks = [...]PreflightCheck{
	{
		skipConfigName:   cmdConfig.SkipCheckVirtEnabled.Name,
		warnConfigName:   cmdConfig.WarnCheckVirtEnabled.Name,
		checkDescription: "Checking if Virtualization is enabled",
		check:            checkVirtualizationEnabled,
		fixDescription:   "Setting up virtualization",
		fix:              fixVirtualizationEnabled,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckKvmEnabled.Name,
		warnConfigName:   cmdConfig.WarnCheckKvmEnabled.Name,
		checkDescription: "Checking if KVM is enabled",
		check:            checkKvmEnabled,
		fixDescription:   "Setting up KVM",
		fix:              fixKvmEnabled,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckLibvirtInstalled.Name,
		warnConfigName:   cmdConfig.WarnCheckLibvirtInstalled.Name,
		checkDescription: "Checking if libvirt is installed",
		check:            checkLibvirtInstalled,
		fixDescription:   "Installing libvirt service and dependencies",
		fix:              fixLibvirtInstalled,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckUserInLibvirtGroup.Name,
		warnConfigName:   cmdConfig.WarnCheckUserInLibvirtGroup.Name,
		checkDescription: "Checking if user is part of libvirt group",
		check:            checkUserPartOfLibvirtGroup,
		fixDescription:   "Adding user to libvirt group",
		fix:              fixUserPartOfLibvirtGroup,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckLibvirtEnabled.Name,
		warnConfigName:   cmdConfig.WarnCheckLibvirtEnabled.Name,
		checkDescription: "Checking if libvirt is enabled",
		check:            checkLibvirtEnabled,
		fixDescription:   "Enabling libvirt",
		fix:              fixLibvirtEnabled,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckLibvirtRunning.Name,
		warnConfigName:   cmdConfig.WarnCheckLibvirtRunning.Name,
		checkDescription: "Checking if libvirt daemon is running",
		check:            checkLibvirtServiceRunning,
		fixDescription:   "Starting libvirt service",
		fix:              fixLibvirtServiceRunning,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckLibvirtVersionCheck.Name,
		warnConfigName:   cmdConfig.WarnCheckLibvirtVersionCheck.Name,
		checkDescription: "Checking if a supported libvirt version is installed",
		check:            checkLibvirtVersion,
		fix:              fixLibvirtVersion,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckLibvirtDriver.Name,
		warnConfigName:   cmdConfig.WarnCheckLibvirtDriver.Name,
		checkDescription: "Checking if crc-driver-libvirt is installed",
		check:            checkMachineDriverLibvirtInstalled,
		fixDescription:   "Installing crc-driver-libvirt",
		fix:              fixMachineDriverLibvirtInstalled,
	},
	{
		check:          checkOldMachineDriverLibvirtInstalled,
		fixDescription: "Removing older system-wide crc-driver-libvirt",
		fix:            fixOldMachineDriverLibvirtInstalled,
		flags:          SetupOnly,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckCrcNetwork.Name,
		warnConfigName:   cmdConfig.WarnCheckCrcNetwork.Name,
		checkDescription: "Checking if libvirt 'crc' network is available",
		check:            checkLibvirtCrcNetworkAvailable,
		fixDescription:   "Setting up libvirt 'crc' network",
		fix:              fixLibvirtCrcNetworkAvailable,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckCrcNetworkActive.Name,
		warnConfigName:   cmdConfig.WarnCheckCrcNetworkActive.Name,
		checkDescription: "Checking if libvirt 'crc' network is active",
		check:            checkLibvirtCrcNetworkActive,
		fixDescription:   "Starting libvirt 'crc' network",
		fix:              fixLibvirtCrcNetworkActive,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckNetworkManagerInstalled.Name,
		warnConfigName:   cmdConfig.WarnCheckNetworkManagerInstalled.Name,
		checkDescription: "Checking if NetworkManager is installed",
		check:            checkNetworkManagerInstalled,
		fixDescription:   "Checking if NetworkManager is installed",
		fix:              fixNetworkManagerInstalled,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckNetworkManagerRunning.Name,
		warnConfigName:   cmdConfig.WarnCheckNetworkManagerRunning.Name,
		checkDescription: "Checking if NetworkManager service is running",
		check:            CheckNetworkManagerIsRunning,
		fixDescription:   "Checking if NetworkManager service is running",
		fix:              fixNetworkManagerIsRunning,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckCrcNetworkManagerConfig.Name,
		warnConfigName:   cmdConfig.WarnCheckCrcNetworkManagerConfig.Name,
		checkDescription: "Checking if /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf exists",
		check:            checkCrcNetworkManagerConfig,
		fixDescription:   "Writing Network Manager config for crc",
		fix:              fixCrcNetworkManagerConfig,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckCrcDnsmasqFile.Name,
		warnConfigName:   cmdConfig.WarnCheckCrcDnsmasqFile.Name,
		checkDescription: "Checking if /etc/NetworkManager/dnsmasq.d/crc.conf exists",
		check:            checkCrcDnsmasqConfigFile,
		fixDescription:   "Writing dnsmasq config for crc",
		fix:              fixCrcDnsmasqConfigFile,
	},
}

func getPreflightChecks() []PreflightCheck {
	checks := []PreflightCheck{}
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, libvirtPreflightChecks[:]...)

	return checks
}

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	doPreflightChecks(getPreflightChecks())
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	doFixPreflightChecks(getPreflightChecks())
}
