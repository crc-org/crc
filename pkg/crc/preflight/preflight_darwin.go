package preflight

import (
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks() {
	preflightCheckSucceedsOrFails(false,
		checkOcBinaryCached,
		"Checking if oc binary is cached",
		false,
	)
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckVirtualBoxInstalled.Name),
		checkVirtualBoxInstalled,
		"Checking if VirtualBox is Installed",
		config.GetBool(cmdConfig.WarnCheckVirtualBoxInstalled.Name),
	)

	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckResolverFilePermissions.Name),
		checkResolverFilePermissions,
		"Checking file permissions for resolver",
		config.GetBool(cmdConfig.WarnCheckResolverFilePermissions.Name),
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost() {
	preflightCheckAndFix(false,
		checkOcBinaryCached,
		fixOcBinaryCached,
		"Caching oc binary",
		false,
	)
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckVirtualBoxInstalled.Name),
		checkVirtualBoxInstalled,
		fixVirtualBoxInstallation,
		"Setting up virtualization",
		config.GetBool(cmdConfig.WarnCheckVirtualBoxInstalled.Name),
	)

	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckResolverFilePermissions.Name),
		checkResolverFilePermissions,
		fixResolverFilePermissions,
		"Setting file permissions for resolver",
		config.GetBool(cmdConfig.WarnCheckResolverFilePermissions.Name),
	)

	// fix user permissions for /etc/resolver/testing
}
