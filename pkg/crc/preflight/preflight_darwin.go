package preflight

import (
	"fmt"
)

// SetupHost performs the prerequisite checks and setups the host to run the cluster
var hyperkitPreflightChecks = [...]Check{
	{
		order:            30,
		configKeySuffix:  "check-hyperkit-installed",
		checkDescription: "Checking if HyperKit is installed",
		check:            checkHyperKitInstalled,
		fixDescription:   "Setting up virtualization with HyperKit",
		fix:              fixHyperKitInstallation,
	},
	{
		order:            30,
		configKeySuffix:  "check-hyperkit-driver",
		checkDescription: "Checking if crc-driver-hyperkit is installed",
		check:            checkMachineDriverHyperKitInstalled,
		fixDescription:   "Installing crc-machine-hyperkit",
		fix:              fixMachineDriverHyperKitInstalled,
	},
}

var dnsPreflightChecks = [...]Check{
	{
		order:            40,
		configKeySuffix:  "check-hosts-file-permissions",
		checkDescription: fmt.Sprintf("Checking file permissions for %s", hostsFile),
		check:            checkEtcHostsFilePermissions,
		fixDescription:   fmt.Sprintf("Setting file permissions for %s", hostsFile),
		fix:              fixEtcHostsFilePermissions,
	},
	{
		order:              40,
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
		order:            200,
		checkDescription: "Checking if tray executable is installed",
		check:            checkTrayExecutablePresent,
		fixDescription:   "Installing and setting up tray",
		fix:              fixTrayExecutablePresent,
		flags:            SetupOnly,
	},
	{
		order:              201,
		checkDescription:   "Checking if launchd configuration for daemon exists",
		check:              checkIfDaemonPlistFileExists,
		fixDescription:     "Creating launchd configuration for daemon",
		fix:                fixDaemonPlistFileExists,
		flags:              SetupOnly,
		cleanupDescription: "Removing launchd configuration for daemon",
		cleanup:            removeDaemonPlistFile,
	},
	{
		order:              202,
		checkDescription:   "Checking if launchd configuration for tray exists",
		check:              checkIfTrayPlistFileExists,
		fixDescription:     "Creating launchd configuration for tray",
		fix:                fixTrayPlistFileExists,
		flags:              SetupOnly,
		cleanupDescription: "Removing launchd configuration for tray",
		cleanup:            removeTrayPlistFile,
	},
	{
		order:            203,
		checkDescription: "Checking installed tray version",
		check:            checkTrayVersion,
		fixDescription:   "Installing and setting up tray app",
		fix:              fixTrayVersion,
		flags:            SetupOnly,
	},
	{
		order:              204,
		checkDescription:   "Checking if CodeReady Containers daemon is running",
		check:              checkIfDaemonAgentRunning,
		fixDescription:     "Starting CodeReady Containers daemon",
		fix:                fixDaemonAgentRunning,
		flags:              SetupOnly,
		cleanupDescription: "Unload CodeReady Containers daemon",
		cleanup:            unLoadDaemonAgent,
	},
	{
		order:              205,
		checkDescription:   "Check if CodeReady Containers tray is running",
		check:              checkIfTrayAgentRunning,
		fixDescription:     "Starting CodeReady Containers tray",
		fix:                fixTrayAgentRunning,
		flags:              SetupOnly,
		cleanupDescription: "Unload CodeReady Containers tray",
		cleanup:            unLoadTrayAgent,
	},
}

func getPreflightChecks(experimentalFeatures bool) []Check {
	checks := []Check{}

	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, nonWinPreflightChecks[:]...)
	checks = append(checks, hyperkitPreflightChecks[:]...)
	checks = append(checks, dnsPreflightChecks[:]...)

	// Experimental feature
	if experimentalFeatures {
		checks = append(checks, traySetupChecks[:]...)
	}

	sort(checks)
	return checks
}
