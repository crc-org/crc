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

var traySetupCheck = PreflightCheck{
	configKeySuffix:  "check-tray-setup",
	checkDescription: "Checking if the tray is installed and running",
	check:            checkTrayExistsAndRunning,
	fixDescription:   "Installing and setting up tray app",
	fix:              fixTrayExistsAndRunning,
	flags:            SetupOnly,
}

func getPreflightChecks() []PreflightCheck {
	checks := []PreflightCheck{}

	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, nonWinPreflightChecks[:]...)
	checks = append(checks, hyperkitPreflightChecks[:]...)
	checks = append(checks, dnsPreflightChecks[:]...)
	checks = append(checks, traySetupCheck)

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
