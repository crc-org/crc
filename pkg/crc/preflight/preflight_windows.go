package preflight

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/adminhelper"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
	"github.com/code-ready/crc/pkg/os/windows/win32"
)

var hypervPreflightChecks = []Check{
	{
		configKeySuffix:  "check-administrator-user",
		checkDescription: "Checking if running in a shell with administrator rights",
		check:            checkIfRunningAsNormalUser,
		fixDescription:   "crc should be ran in a shell without administrator rights",
		flags:            NoFix,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-windows-version",
		checkDescription: "Checking Windows 10 release",
		check:            checkVersionOfWindowsUpdate,
		fixDescription:   "Please manually update your Windows 10 installation",
		flags:            NoFix,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-windows-edition",
		checkDescription: "Checking Windows edition",
		check:            checkWindowsEdition,
		fixDescription:   "Your Windows edition is not supported. Consider using Professional or Enterprise editions of Windows",
		flags:            NoFix,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-hyperv-installed",
		checkDescription: "Checking if Hyper-V is installed and operational",
		check:            checkHyperVInstalled,
		fixDescription:   "Installing Hyper-V",
		fix:              fixHyperVInstalled,

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
		fixDescription: "Creating crc-users group",
		fix: func() error {
			if _, _, err := powershell.ExecuteAsAdmin("create crc-users group", "New-LocalGroup -Name crc-users"); err != nil {
				return fmt.Errorf("failed to create 'crc-users' group: %v", err)
			}
			return nil
		},
		cleanupDescription: "Removing 'crc-users' group",
		cleanup: func() error {
			if _, _, err := powershell.ExecuteAsAdmin("remove crc-users group", "Remove-LocalGroup -Name crc-users"); err != nil {
				return fmt.Errorf("failed to remove 'crc-users' group: %v", err)
			}
			return nil
		},
		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-user-in-hyperv-group",
		checkDescription: "Checking if current user is in Hyper-V group and crc-users group",
		check: func() error {
			if err := adminHelperGroup.check(); err != nil {
				return err
			}
			return hypervGroup.check()
		},
		fixDescription: "Adding current user to group",
		fix: func() error {
			m := crcerrors.MultiError{}
			if err := adminHelperGroup.check(); err != nil {
				m.Collect(adminHelperGroup.fix())
			}
			if err := hypervGroup.check(); err != nil {
				m.Collect(hypervGroup.fix())
			}
			if len(m.Errors) == 0 {
				return nil
			}
			if m.Errors[0] == errReboot {
				return errReboot
			}
			return m
		},
		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-hyperv-service-running",
		checkDescription: "Checking if Hyper-V service is enabled",
		check:            checkHyperVServiceRunning,
		fixDescription:   "Enabling Hyper-V service",
		fix:              fixHyperVServiceRunning,

		labels: labels{Os: Windows},
	},
	{
		configKeySuffix:  "check-hyperv-switch",
		checkDescription: "Checking if the Hyper-V virtual switch exists",
		check:            checkIfHyperVVirtualSwitchExists,
		fixDescription:   "Unable to perform Hyper-V administrative commands. Please reboot your system and run 'crc setup' to complete the setup process",
		flags:            NoFix,

		labels: labels{Os: Windows},
	},
	{
		cleanupDescription: "Removing dns server from interface",
		cleanup:            removeDNSServerAddress,
		flags:              CleanUpOnly,

		labels: labels{Os: Windows},
	},
	{
		cleanupDescription: "Removing the crc VM if exists",
		cleanup:            removeCrcVM,
		flags:              CleanUpOnly,

		labels: labels{Os: Windows},
	},
}

var traySetupChecks = []Check{
	{
		checkDescription:   "Checking if CodeReady Containers daemon is installed",
		check:              checkIfDaemonInstalled,
		fixDescription:     "Removing CodeReady Containers daemon",
		fix:                removeDaemon,
		cleanupDescription: "Uninstalling daemon if installed",
		cleanup:            removeDaemon,
		flags:              SetupOnly,

		labels: labels{Os: Windows, Tray: Enabled},
	},
	{
		checkDescription: "Checking if tray executable is present",
		check:            checkTrayExecutableExists,
		fixDescription:   "Caching tray executable",
		fix:              fixTrayExecutableExists,
		flags:            SetupOnly,

		labels: labels{Os: Windows, Tray: Enabled},
	},
	{
		checkDescription: "Checking if tray is running",
		check:            checkIfTrayRunning,
		fixDescription:   "Starting CodeReady Containers tray",
		fix:              startTray,
		flags:            SetupOnly,

		labels: labels{Os: Windows, Tray: Enabled},
	},
	{
		cleanupDescription: "Stopping tray if running",
		cleanup:            stopTray,
		flags:              CleanUpOnly,

		labels: labels{Os: Windows, Tray: Enabled},
	},
}

var vsockChecks = []Check{
	{
		configKeySuffix:    "check-vsock",
		checkDescription:   "Checking if vsock is correctly configured",
		check:              checkVsock,
		fixDescription:     "Checking if vsock is correctly configured",
		fix:                fixVsock,
		cleanupDescription: "Removing vsock service from hyperv registry",
		cleanup:            cleanVsock,

		labels: labels{Os: Windows, NetworkMode: User},
	},
}

var errReboot = errors.New("Please reboot your system and run 'crc setup' to complete the setup process")

var adminHelperGroup = Check{
	check: func() error {
		_, _, err := powershell.Execute(fmt.Sprintf("Get-LocalGroupMember -Name crc-users -Member '%s'", username()))
		return err
	},
	fix: func() error {
		_, _, err := powershell.ExecuteAsAdmin("adding current user to crc-users group", fmt.Sprintf("Add-LocalGroupMember -Group crc-users -Member '%s'", username()))
		if err != nil {
			return err
		}
		return errReboot
	},
}

func username() string {
	if ok, _ := win32.DomainJoined(); ok {
		return fmt.Sprintf(`%s\%s`, os.Getenv("USERDOMAIN"), os.Getenv("USERNAME"))
	}
	return os.Getenv("USERNAME")
}

var hypervGroup = Check{
	check: checkIfUserPartOfHyperVAdmins,
	fix:   fixUserPartOfHyperVAdmins,
}

var adminHelperChecks = []Check{
	{
		configKeySuffix:  "check-admin-helper-daemon-installed",
		checkDescription: "Checking if admin-helper daemon is installed",
		check: func() error {
			version, err := adminhelper.Client().Version()
			if err != nil {
				return err
			}
			expected := filepath.Base(constants.DefaultAdminHelperCliBase)
			if version != expected {
				return fmt.Errorf("unexpected admin-helper daemon version (expected: %s, got: %s)", expected, version)
			}
			return nil
		},
		fixDescription: "Installing admin-helper daemon",
		fix: func() error {
			_, _, err := powershell.ExecuteAsAdmin("install admin-helper daemon", fmt.Sprintf("& '%s' uninstall-daemon; & '%s' install-daemon", adminhelper.BinPath, adminhelper.BinPath))
			return err
		},
		cleanupDescription: "Uninstalling admin-helper daemon",
		cleanup: func() error {
			_, _, err := powershell.ExecuteAsAdmin("uninstall admin-helper daemon", fmt.Sprintf("& '%s' uninstall-daemon", adminhelper.BinPath))
			return err
		},
		labels: labels{Os: Windows},
	},
	cleanUpHostsFile,
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

func cleanVsock() error {
	if err := checkVsock(); err != nil {
		return nil
	}
	_, _, err := powershell.Execute(fmt.Sprintf(`Remove-Item -Path "%s\%s"`, registryDirectory, registryKey))
	if err != nil {
		return fmt.Errorf("Unable to remove vsock service from hyperv registry: %v", err)
	}
	return nil
}

const (
	Tray LabelName = iota + lastLabelName
)

const (
	// tray
	Enabled LabelValue = iota + lastLabelValue
	Disabled
)

func (filter preflightFilter) SetTray(enable bool) {
	if version.IsMsiBuild() && enable {
		filter[Tray] = Enabled
	} else {
		filter[Tray] = Disabled
	}
}

// We want all preflight checks including
// - experimental checks
// - tray checks when using an installer, regardless of tray enabled or not
// - both user and system networking checks
//
// Passing 'UserNetworkingMode' to getPreflightChecks currently achieves this
// as there are no system networking specific checks
func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, true, network.UserNetworkingMode)
}

func getChecks() []Check {
	checks := []Check{}
	checks = append(checks, genericPreflightChecks...)
	checks = append(checks, hypervPreflightChecks...)
	checks = append(checks, adminHelperChecks...)
	checks = append(checks, vsockChecks...)
	checks = append(checks, traySetupChecks...)
	checks = append(checks, bundleCheck)
	return checks
}

func getPreflightChecks(_ bool, trayAutoStart bool, networkMode network.Mode) []Check {
	filter := newFilter()
	filter.SetNetworkMode(networkMode)
	filter.SetTray(trayAutoStart)

	return filter.Apply(getChecks())
}
