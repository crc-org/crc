package preflight

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	crcpreset "github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
	"github.com/code-ready/crc/pkg/os/windows/win32"
)

var hypervPreflightChecks = []Check{
	{
		configKeySuffix:  "check-administrator-user",
		checkDescription: "Checking if running in a shell with administrator rights",
		check:            checkIfRunningAsNormalUser,
		flags:            StartUpOnly,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-windows-version",
		checkDescription: "Checking Windows 10 release",
		check:            checkVersionOfWindowsUpdate,
		flags:            StartUpOnly,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-windows-edition",
		checkDescription: "Checking Windows edition",
		check:            checkWindowsEdition,
		flags:            StartUpOnly,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-hyperv-installed",
		checkDescription: "Checking if Hyper-V is installed and operational",
		check:            checkHyperVInstalled,
		flags:            StartUpOnly,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-crc-users-group-exists",
		checkDescription: "Checking if crc-users group exists",
		check: func() error {
			if _, _, err := powershell.Execute("Get-LocalGroup -Name crc-users"); err != nil {
				return fmt.Errorf("'crc-users' group does not exist: %v", err)
			}
			return nil
		},
		flags: StartUpOnly,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-user-in-hyperv-group",
		checkDescription: "Checking if current user is in Hyper-V Admins group",
		check:            checkIfUserPartOfHyperVAdmins,
		fixDescription:   "Adding current user to Hyper-V Admins group",
		fix:              fixUserPartOfHyperVAdmins,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-hyperv-service-running",
		checkDescription: "Checking if Hyper-V service is enabled",
		check:            checkHyperVServiceRunning,
		flags:            StartUpOnly,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-hyperv-switch",
		checkDescription: "Checking if the Hyper-V virtual switch exists",
		check:            checkIfHyperVVirtualSwitchExists,
		flags:            StartUpOnly,

		labels: labels{Os: Windows},
	},
	{
		cleanupDescription: "Removing dns server from interface",
		cleanup:            removeDNSServerAddress,
		flags:              CleanUpOnly,

		labels: labels{Os: Windows},
	},
	{
		cleanupDescription: "Removing crc's virtual machine",
		cleanup:            removeCrcVM,
		flags:              CleanUpOnly,

		labels: labels{Os: Windows},
	},
}

var vsockChecks = []Check{
	{
		configKeySuffix:  "check-vsock",
		checkDescription: "Checking if vsock is correctly configured",
		check:            checkVsock,
		flags:            StartUpOnly,

		labels: labels{Os: Windows, NetworkMode: User},
	},
}

var errReboot = errors.New("Please reboot your system and run 'crc setup' to complete the setup process")

func username() string {
	if ok, _ := win32.DomainJoined(); ok {
		return fmt.Sprintf(`%s\%s`, os.Getenv("USERDOMAIN"), os.Getenv("USERNAME"))
	}
	return os.Getenv("USERNAME")
}

const (
	// This key is required to activate the vsock communication
	registryDirectory = `HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices`
	// First part of the key is the vsock port. The rest is not used and just a placeholder.
	registryKey   = "00000400-FACB-11E6-BD58-64006A7986D3"
	registryValue = "gvisor-tap-vsock"
)

func checkVsock() error {
	stdout, _, err := powershell.Execute(fmt.Sprintf(`Get-Item -Path "%s\%s"`, registryDirectory, registryKey))
	if err != nil {
		return err
	}
	if !strings.Contains(stdout, registryValue) {
		return errors.New("VSock registry key not correctly configured")
	}
	return nil
}

// We want all preflight checks including
// - experimental checks
// - tray checks when using an installer, regardless of tray enabled or not
// - both user and system networking checks
//
// Passing 'UserNetworkingMode' to getPreflightChecks currently achieves this
// as there are no system networking specific checks
func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, true, network.UserNetworkingMode, constants.DefaultBundlePath, "")
}

func getChecks(bundlePath string, preset crcpreset.Preset) []Check {
	checks := []Check{}
	checks = append(checks, hypervPreflightChecks...)
	checks = append(checks, vsockChecks...)
	checks = append(checks, bundleCheck(bundlePath))
	checks = append(checks, genericCleanupChecks...)
	return checks
}

func getPreflightChecks(_ bool, trayAutoStart bool, networkMode network.Mode, bundlePath string, preset crcpreset.Preset) []Check {
	filter := newFilter()
	filter.SetNetworkMode(networkMode)

	return filter.Apply(getChecks(bundlePath, preset))
}
