package preflight

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/network"
)

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func hyperkitPreflightChecks(networkMode network.Mode) []Check {
	return []Check{
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

var traySetupChecks = [...]Check{
	{
		checkDescription: "Checking if tray executable is installed",
		check:            checkTrayExecutablePresent,
		fixDescription:   "Installing and setting up tray",
		fix:              fixTrayExecutablePresent,
		flags:            SetupOnly,
	},
	{
		checkDescription:   "Checking if launchd configuration for daemon exists",
		check:              checkIfDaemonPlistFileExists,
		fixDescription:     "Creating launchd configuration for daemon",
		fix:                fixDaemonPlistFileExists,
		flags:              SetupOnly,
		cleanupDescription: "Removing launchd configuration for daemon",
		cleanup:            removeDaemonPlistFile,
	},
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
		checkDescription: "Checking installed tray version",
		check:            checkTrayVersion,
		fixDescription:   "Installing and setting up tray app",
		fix:              fixTrayVersion,
		flags:            SetupOnly,
	},
	{
		checkDescription:   "Checking if CodeReady Containers daemon is running",
		check:              checkIfDaemonAgentRunning,
		fixDescription:     "Starting CodeReady Containers daemon",
		fix:                fixDaemonAgentRunning,
		flags:              SetupOnly,
		cleanupDescription: "Unload CodeReady Containers daemon",
		cleanup:            unLoadDaemonAgent,
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
	return getPreflightChecks(true, network.DefaultMode)
}

func getPreflightChecks(experimentalFeatures bool, mode network.Mode) []Check {
	checks := []Check{}

	checks = append(checks, nonWinPreflightChecks[:]...)
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, hyperkitPreflightChecks(mode)...)
	checks = append(checks, dnsPreflightChecks[:]...)

	if mode == network.DefaultMode {
		checks = append(checks, resolverPreflightChecks[:]...)
	}
	// Experimental feature
	if experimentalFeatures {
		checks = append(checks, traySetupChecks[:]...)
	}

	return checks
}
