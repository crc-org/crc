package preflight

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strings"
	"syscall"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/crc/pkg/os/linux"
)

func libvirtPreflightChecks(distro *linux.OsRelease) []Check {
	checks := []Check{
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
			fix:              fixLibvirtInstalled(distro),
		},
		{
			configKeySuffix:  "check-user-in-libvirt-group",
			checkDescription: "Checking if user is part of libvirt group",
			check:            checkUserPartOfLibvirtGroup,
			fixDescription:   "Adding user to libvirt group",
			fix:              fixUserPartOfLibvirtGroup(distro),
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
			cleanupDescription: "Removing the crc VM if exists",
			cleanup:            removeCrcVM,
			flags:              CleanUpOnly,
		},
	}
	if distroID(distro) == linux.Ubuntu {
		checks = append(checks, ubuntuPreflightChecks...)
	}
	return checks
}

var libvirtNetworkPreflightChecks = [...]Check{
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
}

var vsockPreflightChecks = Check{
	configKeySuffix:  "check-vsock",
	checkDescription: "Checking if vsock is correctly configured",
	check:            checkVsock,
	fixDescription:   "Checking if vsock is correctly configured",
	fix:              fixVsock,
}

func checkVsock() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	getcap, _, err := crcos.RunWithDefaultLocale("getcap", executable)
	if err != nil {
		return err
	}
	if !strings.Contains(string(getcap), "cap_net_bind_service+eip") {
		return fmt.Errorf("capabilities are not correct for %s", executable)
	}
	info, err := os.Stat("/dev/vsock")
	if err != nil {
		return err
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		group, err := user.LookupGroupId(fmt.Sprint(stat.Gid))
		if err != nil {
			return err
		}
		if group.Name != "libvirt" {
			return errors.New("/dev/vsock is not is the right group")
		}
	} else {
		return errors.New("cannot cast info")
	}
	if info.Mode()&0060 == 0 {
		return errors.New("/dev/vsock doesn't have the right permissions")
	}
	return nil
}

func fixVsock() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	_, _, err = crcos.RunWithPrivilege("setcap cap_net_bind_service=+eip", "setcap", "cap_net_bind_service=+eip", executable)
	if err != nil {
		return err
	}
	_, _, err = crcos.RunWithPrivilege("modprobe vhost_vsock", "modprobe", "vhost_vsock")
	if err != nil {
		return err
	}
	_, _, err = crcos.RunWithPrivilege("chown /dev/vsock", "chown", "root:libvirt", "/dev/vsock")
	if err != nil {
		return err
	}
	_, _, err = crcos.RunWithPrivilege("chmod /dev/vsock", "chmod", "g+rw", "/dev/vsock")
	if err != nil {
		return err
	}
	return nil
}

func getAllPreflightChecks() []Check {
	checks := getPreflightChecksForDistro(distro(), network.DefaultMode)
	checks = append(checks, vsockPreflightChecks)
	return checks
}

func getPreflightChecks(_ bool, networkMode network.Mode) []Check {
	return getPreflightChecksForDistro(distro(), networkMode)
}

func getNetworkChecksForDistro(distro *linux.OsRelease, networkMode network.Mode) []Check {
	var checks []Check

	if networkMode == network.VSockMode {
		return append(checks, vsockPreflightChecks)
	}

	switch distroID(distro) {
	default:
		logging.Warnf("distribution-specific preflight checks are not implemented for '%s'", distroID(distro))
		fallthrough
	case linux.RHEL, linux.CentOS, linux.Fedora:
		checks = append(checks, nmPreflightChecks[:]...)
		if usesSystemdResolved(distro) {
			checks = append(checks, systemdResolvedPreflightChecks[:]...)
		} else {
			checks = append(checks, dnsmasqPreflightChecks[:]...)
		}
	case linux.Ubuntu:
		break
	}

	return checks
}

func getPreflightChecksForDistro(distro *linux.OsRelease, networkMode network.Mode) []Check {
	var checks []Check
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, nonWinPreflightChecks[:]...)
	checks = append(checks, libvirtPreflightChecks(distro)...)
	networkChecks := getNetworkChecksForDistro(distro, networkMode)
	checks = append(checks, networkChecks...)
	checks = append(checks, libvirtNetworkPreflightChecks[:]...)
	return checks
}

func usesSystemdResolved(osRelease *linux.OsRelease) bool {
	switch distroID(osRelease) {
	case linux.Fedora:
		return osRelease.VersionID >= "33"
	default:
		return false
	}
}

func distroID(osRelease *linux.OsRelease) linux.OsType {
	if osRelease == nil {
		return "unknown"
	}
	// FIXME: should also use IDLike
	return osRelease.ID
}

func distro() *linux.OsRelease {
	distro, err := linux.GetOsRelease()
	if err != nil {
		logging.Warnf("cannot get distribution name: %v", err)
		return nil
	}
	return distro
}
