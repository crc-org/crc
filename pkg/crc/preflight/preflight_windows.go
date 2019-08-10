package preflight

import ()

var genericPreflightChecks = [...]PreflightCheck{
	{
		checkDescription: "Caching oc binary",
		check:            checkOcBinaryCached,
		fixDescription:   "",
		fix:              fixOcBinaryCached,
	},
	{
		configKeySuffix:  "check-bundle-cached",
		checkDescription: "Unpacking bundle from the CRC binary",
		check:            checkBundleCached,
		fix:              fixBundleCached,
		flags:            SetupOnly,
	},
}

var hypervPreflightChecks = [...]PreflightCheck{
	{
		configKeySuffix:  "check-administrator-user",
		checkDescription: "Checking if running as normal user",
		check:            checkIfRunningAsNormalUserInWindows,
		fix:              fixRunAsNormalUserInWindows,
	},
	{
		configKeySuffix:  "check-windows-version",
		checkDescription: "Check Windows 10 release",
		check:            checkVersionOfWindowsUpdate,
		fix:              fixVersionOfWindowsUpdate,
	},
	{
		configKeySuffix:  "check-hyperv-installed",
		checkDescription: "Hyper-V installed and operational",
		check:            checkHyperVInstalled,
		fix:              fixHyperVInstalled,
	},
	{
		configKeySuffix:  "check-user-in-hyperv-group",
		checkDescription: "Is user a member of the Hyper-V Administrators group",
		check:            checkIfUserPartOfHyperVAdmins,
		fix:              fixUserPartOfHyperVAdmins,
	},
	{
		configKeySuffix:  "check-hyperv-service-running",
		checkDescription: "Hyper-V service enabled",
		check:            checkHyperVServiceRunning,
		fixDescription:   "Hyper-V service enabled",
		fix:              fixHyperVServiceRunning,
	},
	{
		configKeySuffix:  "check-hyperv-switch",
		checkDescription: "Does the Hyper-V virtual switch exist",
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
