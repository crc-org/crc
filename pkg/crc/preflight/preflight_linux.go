package preflight

import (
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os/linux"
)

var libvirtPreflightChecks = [...]Check{
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
		fixDescription:   "Installing a supported libvirt version",
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
		configKeySuffix:  "check-obsolete-libvirt-driver",
		checkDescription: "Checking for obsolete crc-driver-libvirt",
		check:            checkOldMachineDriverLibvirtInstalled,
		fixDescription:   "Removing older system-wide crc-driver-libvirt",
		fix:              fixOldMachineDriverLibvirtInstalled,
		flags:            SetupOnly,
	},
	{
		configKeySuffix:    "check-crc-network",
		checkDescription:   "Checking if libvirt 'crc' network is available",
		check:              checkLibvirtCrcNetworkAvailable,
		fixDescription:     "Setting up libvirt 'crc' network",
		fix:                fixLibvirtCrcNetworkAvailable,
		cleanupDescription: "Removing 'crc' network from libvirt",
		cleanup:            removeLibvirtCrcNetwork,
	},
	{
		configKeySuffix:  "check-crc-network-active",
		checkDescription: "Checking if libvirt 'crc' network is active",
		check:            checkLibvirtCrcNetworkActive,
		fixDescription:   "Starting libvirt 'crc' network",
		fix:              fixLibvirtCrcNetworkActive,
	},
	{
		cleanupDescription: "Removing the crc VM if exists",
		cleanup:            removeCrcVM,
		flags:              CleanUpOnly,
	},
}

func getPreflightChecks(experimentalFeatures bool) []Check {
	return getPreflightChecksForDistro(distro(), experimentalFeatures)
}

func getPreflightChecksForDistro(distro crcos.OsType, experimentalFeatures bool) []Check {
	var checks []Check
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, nonWinPreflightChecks[:]...)
	checks = append(checks, libvirtPreflightChecks[:]...)

	switch distro {
	case crcos.Ubuntu:
	case crcos.RHEL, crcos.CentOS, crcos.Fedora:
		checks = append(checks, redhatPreflightChecks[:]...)
	default:
		logging.Warnf("distribution-specific preflight checks are not implemented for %s", distro)
		checks = append(checks, redhatPreflightChecks[:]...)
	}

	return checks
}

func distro() crcos.OsType {
	distro, err := crcos.GetOsRelease()
	if err != nil {
		logging.Warnf("cannot get distribution name: %v", err)
		return "unknown"
	}
	return distro.ID
}
