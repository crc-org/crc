package preflight

import ()

var hypervPreflightChecks = [...]PreflightCheck{
	{
		configKeySuffix:  "check-administrator-user",
		checkDescription: "Checking if running as normal user",
		check:            checkIfRunningAsNormalUser,
		fix:              fixRunAsNormalUser,
	},
	{
		configKeySuffix:  "check-windows-version",
		checkDescription: "Checking Windows 10 release",
		check:            checkVersionOfWindowsUpdate,
		fix:              fixVersionOfWindowsUpdate,
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
		fix:              fixHyperVVirtualSwitch,
	},
}

func getPreflightChecks() []PreflightCheck {
	checks := []PreflightCheck{}
	checks = append(checks, genericPreflightChecks[:]...)
	checks = append(checks, hypervPreflightChecks[:]...)

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
