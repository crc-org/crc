package preflight

import (
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckAdministratorUser.Name),
		checkIfRunningAsNormalUserInWindows,
		"Checking if running as normal user",
		config.GetBool(cmdConfig.WarnCheckAdministratorUser.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckWindowsVersionCheck.Name),
		checkOcBinaryCached,
		"Checking if oc binary is cached",
		config.GetBool(cmdConfig.WarnCheckWindowsVersionCheck.Name),
	)

	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckWindowsVersionCheck.Name),
		checkVersionOfWindowsUpdate,
		"Check Windows 10 release",
		config.GetBool(cmdConfig.WarnCheckWindowsVersionCheck.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckHyperVInstalled.Name),
		checkHyperVInstalled,
		"Hyper-V installed and operational",
		config.GetBool(cmdConfig.WarnCheckHyperVInstalled.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckUserInHyperVGroup.Name),
		checkIfUserPartOfHyperVAdmins,
		"Is user a member of the Hyper-V Administrators group",
		config.GetBool(cmdConfig.WarnCheckUserInHyperVGroup.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckHyperVServiceRunning.Name),
		checkHyperVServiceRunning,
		"Hyper-V service enabled",
		config.GetBool(cmdConfig.WarnCheckHyperVServiceRunning.Name),
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckHyperVSwitch.Name),
		checkIfHyperVVirtualSwitchExists,
		"Does the Hyper-V virtual switch exist",
		config.GetBool(cmdConfig.WarnCheckHyperVSwitch.Name),
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckAdministratorUser.Name),
		checkIfRunningAsNormalUserInWindows,
		fixRunAsNormalUserInWindows,
		"Checking if running as normal user",
		config.GetBool(cmdConfig.WarnCheckAdministratorUser.Name),
	)
	preflightCheckAndFix(false,
		checkOcBinaryCached,
		fixOcBinaryCached,
		"Caching oc binary",
		false,
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckBundleCached.Name),
		checkBundleCached,
		fixBundleCached,
		"Unpacking bundle from the CRC binary",
		config.GetBool(cmdConfig.WarnCheckBundleCached.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckWindowsVersionCheck.Name),
		checkVersionOfWindowsUpdate,
		fixVersionOfWindowsUpdate,
		"Check Windows 10 release",
		config.GetBool(cmdConfig.WarnCheckWindowsVersionCheck.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckHyperVInstalled.Name),
		checkHyperVInstalled,
		fixHyperVInstalled,
		"Hyper-V installed",
		config.GetBool(cmdConfig.WarnCheckHyperVInstalled.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckUserInHyperVGroup.Name),
		// Workaround to an issue the check returns "True"
		checkIfUserPartOfHyperVAdmins,
		fixUserPartOfHyperVAdmins,
		"Is user a member of the Hyper-V Administrators group",
		config.GetBool(cmdConfig.WarnCheckUserInHyperVGroup.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckHyperVServiceRunning.Name),
		checkHyperVServiceRunning,
		fixHyperVServiceRunning,
		"Hyper-V service enabled",
		config.GetBool(cmdConfig.WarnCheckHyperVServiceRunning.Name),
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckHyperVSwitch.Name),
		checkIfHyperVVirtualSwitchExists,
		fixHyperVVirtualSwitch,
		"Does the Hyper-V virtual switch exist",
		config.GetBool(cmdConfig.WarnCheckHyperVSwitch.Name),
	)
}
