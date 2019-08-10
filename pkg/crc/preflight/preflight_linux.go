package preflight

import ()

var libvirtPreflightChecks = [...]PreflightCheck{
	{
		configKeySuffix:  "check-virt-enabled",
		checkDescription: "Checking if Virtualization is enabled",
		check:            checkVirtualizationEnabled,
		fixDescription:   "Setting up virtualization",
		fix:              fixVirtualizationEnabled,
	},
	{
		configKeySuffix:  "check-kvm-enabled",
		checkDescription: "Checking if KVM is enabled",
		check:            checkKvmEnabled,
		fixDescription:   "Setting up KVM",
		fix:              fixKvmEnabled,
	},
	{
		configKeySuffix:  "check-libvirt-installed",
		checkDescription: "Checking if libvirt is installed",
		check:            checkLibvirtInstalled,
		fixDescription:   "Installing libvirt service and dependencies",
		fix:              fixLibvirtInstalled,
	},
	{
		configKeySuffix:  "check-user-in-libvirt-group",
		checkDescription: "Checking if user is part of libvirt group",
		check:            checkUserPartOfLibvirtGroup,
		fixDescription:   "Adding user to libvirt group",
		fix:              fixUserPartOfLibvirtGroup,
	},
	{
		configKeySuffix:  "check-libvirt-enabled",
		checkDescription: "Checking if libvirt is enabled",
		check:            checkLibvirtEnabled,
		fixDescription:   "Enabling libvirt",
		fix:              fixLibvirtEnabled,
	},
	{
		configKeySuffix:  "check-libvirt-running",
		checkDescription: "Checking if libvirt daemon is running",
		check:            checkLibvirtServiceRunning,
		fixDescription:   "Starting libvirt service",
		fix:              fixLibvirtServiceRunning,
	},
	{
		configKeySuffix:  "check-libvirt-version",
		checkDescription: "Checking if a supported libvirt version is installed",
		check:            checkLibvirtVersion,
		fix:              fixLibvirtVersion,
	},
	{
		configKeySuffix:  "check-libvirt-driver",
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
		configKeySuffix:  "check-crc-network",
		checkDescription: "Checking if libvirt 'crc' network is available",
		check:            checkLibvirtCrcNetworkAvailable,
		fixDescription:   "Setting up libvirt 'crc' network",
		fix:              fixLibvirtCrcNetworkAvailable,
	},
	{
		configKeySuffix:  "check-crc-network-active",
		checkDescription: "Checking if libvirt 'crc' network is active",
		check:            checkLibvirtCrcNetworkActive,
		fixDescription:   "Starting libvirt 'crc' network",
		fix:              fixLibvirtCrcNetworkActive,
	},
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
		check:            CheckNetworkManagerIsRunning,
		fixDescription:   "Checking if NetworkManager service is running",
		fix:              fixNetworkManagerIsRunning,
	},
	{
		configKeySuffix:  "check-network-manager-config",
		checkDescription: "Checking if /etc/NetworkManager/conf.d/crc-nm-dnsmasq.conf exists",
		check:            checkCrcNetworkManagerConfig,
		fixDescription:   "Writing Network Manager config for crc",
		fix:              fixCrcNetworkManagerConfig,
	},
	{
		configKeySuffix:  "check-crc-dnsmasq-file",
		checkDescription: "Checking if /etc/NetworkManager/dnsmasq.d/crc.conf exists",
		check:            checkCrcDnsmasqConfigFile,
		fixDescription:   "Writing dnsmasq config for crc",
		fix:              fixCrcDnsmasqConfigFile,
	},
}

func getPreflightChecks() []PreflightCheck {
	checks := []PreflightCheck{}
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, nonWinPreflightChecks[:]...)
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

func RegisterSettings() {
	doRegisterSettings(getPreflightChecks())
}
