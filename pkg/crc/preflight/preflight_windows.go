package preflight

import (
	"errors"

	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks(vmDriver string) {
	preflightCheckSucceedsOrFails(false,
		checkIfRunningAsNormalUserInWindows,
		"Checking if running as adminstrator",
		false,
	)
	preflightCheckSucceedsOrFails(false,
		checkOcBinaryCached,
		"Checking if oc binary is cached",
		false,
	)

	if vmDriver == "hyperv" {
		preflightCheckSucceedsOrFails(false,
			checkVersionOfWindowsUpdate,
			"Check Windows 10 release",
			false,
		)
		preflightCheckSucceedsOrFails(false,
			checkHyperVInstalled,
			"Hyper-V installed and operational",
			false,
		)
		preflightCheckSucceedsOrFails(false,
			checkIfUserPartOfHyperVAdmins,
			"Is user a member of the Hyper-V Administrators group",
			false,
		)
		preflightCheckSucceedsOrFails(false,
			checkIfHyperVVirtualSwitchExists,
			"Does the Hyper-V virtual switch exist",
			false,
		)
	}
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost(vmDriver string) {
	preflightCheckAndFix(false,
		checkIfRunningAsNormalUserInWindows,
		fixRunAsNormalUserInWindows,
		"Checking if running as adminstrator",
		false,
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

	if vmDriver == "hyperv" {
		preflightCheckAndFix(false,
			checkVersionOfWindowsUpdate,
			fixVersionOfWindowsUpdate,
			"Check Windows 10 release",
			false,
		)
		preflightCheckAndFix(false,
			checkHyperVInstalled,
			fixHyperVInstalled,
			"Hyper-V installed",
			false,
		)
		preflightCheckAndFix(false,
			// Workaround to an issue the check returns "True"
			func() (bool, error) { return false, errors.New("Always add user") },
			fixUserPartOfHyperVAdmins,
			"Is user a member of the Hyper-V Administrators group",
			false,
		)
		preflightCheckAndFix(false,
			checkIfHyperVVirtualSwitchExists,
			fixHyperVVirtualSwitch,
			"Does the Hyper-V virtual switch exist",
			false,
		)
	}
}
