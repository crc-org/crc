package preflight

import (
	"errors"
	"fmt"
	"strings"

	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

var hypervPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-administrator-user",
		checkDescription: "Checking if running in a shell with administrator rights",
		check:            checkIfRunningAsNormalUser,
		fixDescription:   "crc should be ran in a shell without administrator rights",
		flags:            NoFix,
	},
	{
		configKeySuffix:  "check-windows-version",
		checkDescription: "Checking Windows 10 release",
		check:            checkVersionOfWindowsUpdate,
		fixDescription:   "Please manually update your Windows 10 installation",
		flags:            NoFix,
	},
	{
		configKeySuffix:  "check-windows-edition",
		checkDescription: "Checking Windows edition",
		check:            checkWindowsEdition,
		fixDescription:   "Your Windows edition is not supported. Consider using Professional or Enterprise editions of Windows",
		flags:            NoFix,
	},
	{
		configKeySuffix:  "check-hyperv-installed",
		checkDescription: "Checking if Hyper-V is installed and operational",
		check:            checkHyperVInstalled,
		fixDescription:   "Installing Hyper-V",
		fix:              fixHyperVInstalled,
	},
	{
		configKeySuffix:  "check-user-in-hyperv-group",
		checkDescription: "Checking if user is a member of the Hyper-V Administrators group",
		check:            checkIfUserPartOfHyperVAdmins,
		fixDescription:   "Adding user to the Hyper-V Administrators group",
		fix:              fixUserPartOfHyperVAdmins,
	},
	{
		configKeySuffix:  "check-hyperv-service-running",
		checkDescription: "Checking if Hyper-V service is enabled",
		check:            checkHyperVServiceRunning,
		fixDescription:   "Enabling Hyper-V service",
		fix:              fixHyperVServiceRunning,
	},
	{
		configKeySuffix:  "check-hyperv-switch",
		checkDescription: "Checking if the Hyper-V virtual switch exist",
		check:            checkIfHyperVVirtualSwitchExists,
		fixDescription:   "Unable to perform Hyper-V administrative commands. Please reboot your system and run 'crc setup' to complete the setup process",
		flags:            NoFix,
	},
	{
		cleanupDescription: "Removing dns server from interface",
		cleanup:            removeDNSServerAddress,
		flags:              CleanUpOnly,
	},
	{
		cleanupDescription: "Removing the crc VM if exists",
		cleanup:            removeCrcVM,
		flags:              CleanUpOnly,
	},
}

var traySetupChecks = [...]Check{
	{
		checkDescription: "Checking if tray executable is present",
		check:            checkTrayExecutableExists,
		fixDescription:   "Caching tray executable",
		fix:              fixTrayExecutableExists,
		flags:            SetupOnly,
	},
	{
		checkDescription:   "Checking if CodeReady Containers daemon is installed",
		check:              checkIfDaemonInstalled,
		fixDescription:     "Installing CodeReady Containers daemon",
		fix:                fixDaemonInstalled,
		cleanupDescription: "Uninstalling daemon if installed",
		cleanup:            removeDaemon,
		flags:              SetupOnly,
	},
	{
		checkDescription:   "Checking if tray is installed",
		check:              checkIfTrayInstalled,
		fixDescription:     "Installing CodeReady Containers tray",
		fix:                fixTrayInstalled,
		cleanupDescription: "Uninstalling tray if installed",
		cleanup:            removeTray,
		flags:              SetupOnly,
	},
}

var vsockChecks = [...]Check{
	{
		configKeySuffix:  "check-vsock",
		checkDescription: "Checking if vsock is correctly configured",
		check:            checkVsock,
		fixDescription:   "Checking if vsock is correctly configured",
		fix:              fixVsock,
	},
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

func fixVsock() error {
	cmds := []string{
		fmt.Sprintf(`$service = New-Item -Path "%s" -Name "%s"`, registryDirectory, registryKey),
		fmt.Sprintf(`$service.SetValue("ElementName", "%v")`, registryValue),
	}
	_, _, err := powershell.ExecuteAsAdmin("adding vsock registry key", strings.Join(cmds, ";"))
	return err
}

func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, network.VSockMode)
}

func getPreflightChecks(experimentalFeatures bool, networkMode network.Mode) []Check {
	checks := []Check{}
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, hypervPreflightChecks[:]...)

	if networkMode == network.VSockMode {
		checks = append(checks, vsockChecks[:]...)
	}

	// Experimental feature
	if experimentalFeatures {
		checks = append(checks, traySetupChecks[:]...)
	}

	checks = append(checks, bundleCheck)
	return checks
}
