package preflight

import (
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckVirtualBoxInstalled.Name),
		checkVirtualBoxInstalled,
		"Checking if VirtualBox is Installed",
		config.GetBool(cmdConfig.WarnCheckVirtualBoxInstalled.Name),
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckVirtualBoxInstalled.Name),
		checkVirtualBoxInstalled,
		fixVirtualBoxInstallation,
		"Setting up virtualization",
		config.GetBool(cmdConfig.WarnCheckVirtualBoxInstalled.Name),
	)
}
