package preflight

import (
	"fmt"

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
		},
		{
			configKeySuffix:  "check-hyperkit-installed",
			checkDescription: "Checking if HyperKit is installed",
			check:            checkHyperKitInstalled(networkMode),
			fixDescription:   "Setting up virtualization with HyperKit",
			fix:              fixHyperKitInstallation(networkMode),
		},
		{
			configKeySuffix:  "check-hyperkit-driver",
			checkDescription: "Checking if crc-driver-hyperkit is installed",
			check:            checkMachineDriverHyperKitInstalled(networkMode),
			fixDescription:   "Installing crc-machine-hyperkit",
			fix:              fixMachineDriverHyperKitInstalled(networkMode),
		},
		{
			cleanupDescription: "Stopping CRC Hyperkit process",
			cleanup:            stopCRCHyperkitProcess,
			flags:              CleanUpOnly,
		},
	}
}

var dnsPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-hosts-file-permissions",
		checkDescription: fmt.Sprintf("Checking file permissions for %s", hostsFile),
		check:            checkEtcHostsFilePermissions,
		fixDescription:   fmt.Sprintf("Setting file permissions for %s", hostsFile),
		fix:              fixEtcHostsFilePermissions,
	},
}

var resolverPreflightChecks = [...]Check{
	{
		configKeySuffix:    "check-resolver-file-permissions",
		checkDescription:   fmt.Sprintf("Checking file permissions for %s", resolverFile),
		check:              checkResolverFilePermissions,
		fixDescription:     fmt.Sprintf("Setting file permissions for %s", resolverFile),
		fix:                fixResolverFilePermissions,
		cleanupDescription: fmt.Sprintf("Removing %s file", resolverFile),
		cleanup:            removeResolverFile,
	},
}

var daemonSetupChecks = [...]Check{
	{
		checkDescription:   "Checking if CodeReady Containers daemon is running",
		check:              checkIfDaemonAgentRunning,
		fixDescription:     "Unloading CodeReady Containers daemon",
		fix:                unLoadDaemonAgent,
		flags:              SetupOnly,
		cleanupDescription: "Unload CodeReady Containers daemon",
		cleanup:            unLoadDaemonAgent,
	},
}

var traySetupChecks = [...]Check{
	{
		checkDescription:   "Checking if launchd configuration for tray exists",
		check:              checkIfTrayPlistFileExists,
		fixDescription:     "Creating launchd configuration for tray",
		fix:                fixTrayPlistFileExists,
		flags:              SetupOnly,
		cleanupDescription: "Removing launchd configuration for tray",
		cleanup:            removeTrayPlistFile,
	},
	{
		checkDescription:   "Check if CodeReady Containers tray is running",
		check:              checkIfTrayAgentRunning,
		fixDescription:     "Starting CodeReady Containers tray",
		fix:                fixTrayAgentRunning,
		flags:              SetupOnly,
		cleanupDescription: "Unload CodeReady Containers tray",
		cleanup:            unLoadTrayAgent,
	},
}

func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, true, network.DefaultMode)
}

func getPreflightChecks(experimentalFeatures bool, trayAutostart bool, mode network.Mode) []Check {
	checks := []Check{}

	checks = append(checks, nonWinPreflightChecks[:]...)
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, hyperkitPreflightChecks(mode)...)
	checks = append(checks, dnsPreflightChecks[:]...)
	checks = append(checks, daemonSetupChecks[:]...)

	if mode == network.DefaultMode {
		checks = append(checks, resolverPreflightChecks[:]...)
	}

	if version.IsMacosInstallPathSet() && trayAutostart {
		checks = append(checks, traySetupChecks[:]...)
	}

	checks = append(checks, bundleCheck)
	return checks
}
