package preflight

import (
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
)

var genericPreflightChecks = [...]PreflightCheck{
	{
		checkDescription: "Caching oc binary",
		check:            checkOcBinaryCached,
		fixDescription:   "",
		fix:              fixOcBinaryCached,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckBundleCached.Name,
		warnConfigName:   cmdConfig.WarnCheckBundleCached.Name,
		checkDescription: "Unpacking bundle from the CRC binary",
		check:            checkBundleCached,
		fix:              fixBundleCached,
		flags:            SetupOnly,
	},
}

var hypervPreflightChecks = [...]PreflightCheck{
	{
		skipConfigName:   cmdConfig.SkipCheckAdministratorUser.Name,
		warnConfigName:   cmdConfig.WarnCheckAdministratorUser.Name,
		checkDescription: "Checking if running as normal user",
		check:            checkIfRunningAsNormalUserInWindows,
		fix:              fixRunAsNormalUserInWindows,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckWindowsVersionCheck.Name,
		warnConfigName:   cmdConfig.WarnCheckWindowsVersionCheck.Name,
		checkDescription: "Check Windows 10 release",
		check:            checkVersionOfWindowsUpdate,
		fix:              fixVersionOfWindowsUpdate,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckHyperVInstalled.Name,
		warnConfigName:   cmdConfig.WarnCheckHyperVInstalled.Name,
		checkDescription: "Hyper-V installed and operational",
		check:            checkHyperVInstalled,
		fix:              fixHyperVInstalled,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckUserInHyperVGroup.Name,
		warnConfigName:   cmdConfig.WarnCheckUserInHyperVGroup.Name,
		checkDescription: "Is user a member of the Hyper-V Administrators group",
		check:            checkIfUserPartOfHyperVAdmins,
		fix:              fixUserPartOfHyperVAdmins,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckHyperVServiceRunning.Name,
		warnConfigName:   cmdConfig.WarnCheckHyperVServiceRunning.Name,
		checkDescription: "Hyper-V service enabled",
		check:            checkHyperVServiceRunning,
		fixDescription:   "Hyper-V service enabled",
		fix:              fixHyperVServiceRunning,
	},
	{
		skipConfigName:   cmdConfig.SkipCheckHyperVSwitch.Name,
		warnConfigName:   cmdConfig.WarnCheckHyperVSwitch.Name,
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
