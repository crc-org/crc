package preflight

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/version"
)

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func hyperkitPreflightChecks(networkMode network.Mode) []Check {
	return []Check{
		{
			configKeySuffix:  "check-m1-cpu",
			checkDescription: "Checking if running emulated on a M1 CPU",
			check:            checkM1CPU,
			fixDescription:   "CodeReady Containers is unsupported on Apple M1 hardware",
			flags:            NoFix,

			labels: labels{Os: Darwin},
		},
		{
			configKeySuffix:  "check-hyperkit-installed",
			checkDescription: "Checking if HyperKit is installed",
			check:            checkHyperKitInstalled(networkMode),
			fixDescription:   "Setting up virtualization with HyperKit",
			fix:              fixHyperKitInstallation(networkMode),

			labels: labels{Os: Darwin},
		},
		{
			configKeySuffix:  "check-qcow-tool-installed",
			checkDescription: "Checking if qcow-tool is installed",
			check:            checkQcowToolInstalled,
			fixDescription:   "Installing qcow-tool",
			fix:              fixQcowToolInstalled,

			labels: labels{Os: Darwin},
		},
		{
			configKeySuffix:  "check-hyperkit-driver",
			checkDescription: "Checking if crc-driver-hyperkit is installed",
			check:            checkMachineDriverHyperKitInstalled(networkMode),
			fixDescription:   "Installing crc-machine-hyperkit",
			fix:              fixMachineDriverHyperKitInstalled(networkMode),

			labels: labels{Os: Darwin},
		},
		{
			cleanupDescription: "Stopping CRC Hyperkit process",
			cleanup:            stopCRCHyperkitProcess,
			flags:              CleanUpOnly,

			labels: labels{Os: Darwin},
		},
	}
}

var resolverPreflightChecks = []Check{
	{
		configKeySuffix:    "check-resolver-file-permissions",
		checkDescription:   fmt.Sprintf("Checking file permissions for %s", resolverFile),
		check:              checkResolverFilePermissions,
		fixDescription:     fmt.Sprintf("Setting file permissions for %s", resolverFile),
		fix:                fixResolverFilePermissions,
		cleanupDescription: fmt.Sprintf("Removing %s file", resolverFile),
		cleanup:            removeResolverFile,

		labels: labels{Os: Darwin, NetworkMode: System},
	},
}

var daemonSetupChecks = []Check{
	{
		checkDescription:   "Checking if CodeReady Containers daemon is running",
		check:              checkIfDaemonAgentRunning,
		fixDescription:     "Unloading CodeReady Containers daemon",
		fix:                unLoadDaemonAgent,
		flags:              SetupOnly,
		cleanupDescription: "Unload CodeReady Containers daemon",
		cleanup:            unLoadDaemonAgent,

		labels: labels{Os: Darwin},
	},
}

var traySetupChecks = []Check{
	{
		checkDescription:   "Checking if launchd configuration for tray exists",
		check:              checkIfTrayPlistFileExists,
		fixDescription:     "Creating launchd configuration for tray",
		fix:                fixTrayPlistFileExists,
		flags:              SetupOnly,
		cleanupDescription: "Removing launchd configuration for tray",
		cleanup:            removeTrayPlistFile,

		labels: labels{Os: Darwin, Tray: Enabled},
	},
	{
		checkDescription:   "Check if CodeReady Containers tray is running",
		check:              checkIfTrayAgentRunning,
		fixDescription:     "Starting CodeReady Containers tray",
		fix:                fixTrayAgentRunning,
		flags:              SetupOnly,
		cleanupDescription: "Unload CodeReady Containers tray",
		cleanup:            unLoadTrayAgent,

		labels: labels{Os: Darwin, Tray: Enabled},
	},
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
	if version.IsInstaller() && enable {
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
// Passing 'SystemNetworkingMode' to getPreflightChecks currently achieves this
// as there are no user networking specific checks
func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, true, network.SystemNetworkingMode, constants.DefaultBundlePath, "")
}

func getChecks(mode network.Mode, bundlePath, preset string) []Check {
	checks := []Check{}

	checks = append(checks, nonWinPreflightChecks...)
	checks = append(checks, genericPreflightChecks...)
	checks = append(checks, genericCleanupChecks...)
	checks = append(checks, hyperkitPreflightChecks(mode)...)
	checks = append(checks, daemonSetupChecks...)
	checks = append(checks, resolverPreflightChecks...)
	checks = append(checks, traySetupChecks...)
	checks = append(checks, bundleCheck(bundlePath, preset))

	return checks
}

func getPreflightChecks(_ bool, trayAutostart bool, mode network.Mode, bundlePath, preset string) []Check {
	filter := newFilter()
	filter.SetNetworkMode(mode)
	filter.SetTray(trayAutostart)

	return filter.Apply(getChecks(mode, bundlePath, preset))
}
