package preflight

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/logging"
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
		configKeySuffix:  "check-resolver-file-permissions",
		checkDescription: fmt.Sprintf("Checking file permissions for %s", resolverFile),
		check:            checkResolverFilePermissions,
		fixDescription:   fmt.Sprintf("Setting file permissions for %s", resolverFile),
		fix:              fixResolverFilePermissions,
	},
	{
		configKeySuffix:  "check-hosts-file-permissions",
		checkDescription: fmt.Sprintf("Checking file permissions for %s", hostFile),
		check:            checkHostsFilePermissions,
		fixDescription:   fmt.Sprintf("Setting file permissions for %s", hostFile),
		fix:              fixHostsFilePermissions,
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
		checkDescription: "Checking if launchd configuration for daemon exists",
		check:            checkIfDaemonPlistFileExists,
		fixDescription:   "Creating launchd configuration for daemon",
		fix:              fixDaemonPlistFileExists,
		flags:            SetupOnly,
	},
	{
		checkDescription: "Checking if launchd configuration for tray exists",
		check:            checkIfTrayPlistFileExists,
		fixDescription:   "Creating launchd configuration for tray",
		fix:              fixTrayPlistFileExists,
		flags:            SetupOnly,
	},
	{
		checkDescription: "Checking installed tray version",
		check:            checkTrayVersion,
		fixDescription:   "Installing and setting up tray app",
		fix:              fixTrayVersion,
		flags:            SetupOnly,
	},
	{
		checkDescription: "Checking if CodeReady Containers daemon is running",
		check:            checkIfDaemonAgentRunning,
		fixDescription:   "Starting CodeReady Containers daemon",
		fix:              fixDaemonAgentRunning,
		flags:            SetupOnly,
	},
	{
		checkDescription: "Check if CodeReady Containers tray is running",
		check:            checkIfTrayAgentRunning,
		fixDescription:   "Starting CodeReady Containers tray",
		fix:              fixTrayAgentRunning,
		flags:            SetupOnly,
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
	logging.Warn("Cleanup is not supported for MacOS")
	doCleanUpPreflightChecks(getPreflightChecks())
}
