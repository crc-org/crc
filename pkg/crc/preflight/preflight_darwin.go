package preflight

import (
	"fmt"
)

// SetupHost performs the prerequisite checks and setups the host to run the cluster
var hyperkitPreflightChecks = [...]PreflightCheck{
	{
		configKeySuffix:  "check-hyperkit-installed",
		checkDescription: "Checking if HyperKit is installed",
		check:            checkHyperKitInstalled,
		fixDescription:   "Setting up virtualization with HyperKit",
		fix:              fixHyperKitInstallation,
	},
	{
		configKeySuffix:  "check-hyperkit-driver",
		checkDescription: "Checking if crc-driver-hyperkit is installed",
		check:            checkMachineDriverHyperKitInstalled,
		fixDescription:   "Installing crc-machine-hyperkit",
		fix:              fixMachineDriverHyperKitInstalled,
	},
}

var dnsPreflightChecks = [...]PreflightCheck{
	{
		configKeySuffix:  "check-hosts-file-permissions",
		checkDescription: fmt.Sprintf("Checking file permissions for %s", hostsFile),
		check:            checkEtcHostsFilePermissions,
		fixDescription:   fmt.Sprintf("Setting file permissions for %s", hostsFile),
		fix:              fixEtcHostsFilePermissions,
	},
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

var traySetupChecks = [...]PreflightCheck{
	{
		checkDescription: "Checking if tray binary is installed",
		check:            checkTrayBinaryPresent,
		fixDescription:   "Installing and setting up tray",
		fix:              fixTrayBinaryPresent,
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

func getPreflightChecks() []PreflightCheck {
	checks := []PreflightCheck{}

	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, nonWinPreflightChecks[:]...)
	checks = append(checks, hyperkitPreflightChecks[:]...)
	checks = append(checks, dnsPreflightChecks[:]...)

	// Experimental feature
	if EnableExperimentalFeatures {
		checks = append(checks, traySetupChecks[:]...)
	}

	return checks
}

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	doPreflightChecks(getPreflightChecks())
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	doFixPreflightChecks(getPreflightChecks())
}

func RegisterSettings() {
	doRegisterSettings(getPreflightChecks())
}

func CleanUpHost() {
	// A user can use setup with experiment flag
	// and not use cleanup with same flag, to avoid
	// any extra step/confusion we are just adding the checks
	// which are behind the experiment flag. This way cleanup
	// perform action in a sane way.
	checks := getPreflightChecks()
	checks = append(checks, traySetupChecks[:]...)
	doCleanUpPreflightChecks(checks)
}
