package preflight

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
	"github.com/crc-org/crc/v2/pkg/os/windows/win32"
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
		checkDescription: "Checking Windows release",
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

		labels: labels{Os: Windows, NetworkMode: System},
	},
	{
		cleanupDescription: "Removing dns server from interface",
		cleanup:            removeDNSServerAddress,
		flags:              CleanUpOnly,

		labels: labels{Os: Windows, NetworkMode: System},
	},
}

var cleanupCheckRemoveCrcVM = Check{
	cleanupDescription: "Removing crc's virtual machine",
	cleanup:            removeCrcVM,
	flags:              CleanUpOnly,

	labels: labels{Os: Windows},
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

var daemonTaskChecks = []Check{
	{
		cleanupDescription: "Stop daemon process if it is running",
		cleanup:            killDaemonProcessIfRunning,
		flags:              CleanUpOnly,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:    "check-background-launcher-install",
		checkDescription:   "Checking if the win32 background launcher is installed",
		check:              checkWin32BackgroundLauncherInstalled,
		fixDescription:     "Installing the win32 background launcher",
		fix:                fixWin32BackgroundLauncherInstalled,
		cleanupDescription: "Removing the win32 background launcher",
		cleanup:            removeWin32BackgroundLauncher,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:    "check-daemon-task-install",
		checkDescription:   "Checking if the daemon task is installed",
		check:              checkIfDaemonTaskInstalled,
		fixDescription:     "Installing the daemon task",
		fix:                fixDaemonTaskInstalled,
		cleanupDescription: "Removing the daemon task",
		cleanup:            removeDaemonTask,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-daemon-task-running",
		checkDescription: "Checking if the daemon task is running",
		check:            checkIfDaemonTaskRunning,
		fixDescription:   "Running the daemon task",
		fix:              fixDaemonTaskRunning,

		labels: labels{Os: Windows},
	},
}

var adminHelperServiceCheks = []Check{
	{
		configKeySuffix:  "check-admin-helper-service-running",
		checkDescription: "Checking admin helper service is running",
		check:            checkIfAdminHelperServiceRunning,
		fixDescription:   "Make sure you installed crc using the Windows installer and performed required reboot",
		flags:            NoFix,

		labels: labels{Os: Windows},
	},
}

// 'crc-user' group is supposed to be created by the msi or chocolatey
// this check makes sure that was done before stating crc, it does not
// have a fix function since this'll be handled by the msi or choco
var crcUsersGroupExistsCheck = Check{
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
}

var userPartOfCrcUsersAndHypervAdminsGroupCheck = Check{
	configKeySuffix:  "check-user-in-crc-users-and-hyperv-admins-group",
	checkDescription: "Checking if current user is in crc-users and Hyper-V admins group",
	check:            checkUserPartOfCrcUsersAndHypervAdminsGroup,
	fixDescription:   "Adding logon user to crc-users and Hyper-V admins group",
	fix:              fixUserPartOfCrcUsersAndHypervAdminsGroup,

	labels: labels{Os: Windows},
}

var errReboot = errors.New("Please reboot your system and run 'crc setup' to complete the setup process")

func username() string {
	if ok := win32.DomainJoined(); ok {
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
// - both user and system networking checks
//
// Passing 'UserNetworkingMode' to getPreflightChecks currently achieves this
// as there are no system networking specific checks
func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, network.UserNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift, false)
}

func getChecks(bundlePath string, preset crcpreset.Preset, enableBundleQuayFallback bool) []Check {
	checks := []Check{}
	checks = append(checks, memoryCheck(preset))
	checks = append(checks, removePodmanFromOcBinDirCheck())
	checks = append(checks, hypervPreflightChecks...)
	checks = append(checks, crcUsersGroupExistsCheck)
	checks = append(checks, userPartOfCrcUsersAndHypervAdminsGroupCheck)
	checks = append(checks, vsockChecks...)
	checks = append(checks, bundleCheck(bundlePath, preset, enableBundleQuayFallback))
	checks = append(checks, genericCleanupChecks...)
	checks = append(checks, cleanupCheckRemoveCrcVM)
	checks = append(checks, daemonTaskChecks...)
	checks = append(checks, adminHelperServiceCheks...)
	checks = append(checks, sshPortCheck())
	return checks
}

func getPreflightChecks(_ bool, networkMode network.Mode, bundlePath string, preset crcpreset.Preset, enableBundleQuayFallback bool) []Check {
	filter := newFilter()
	filter.SetNetworkMode(networkMode)

	return filter.Apply(getChecks(bundlePath, preset, enableBundleQuayFallback))
}
