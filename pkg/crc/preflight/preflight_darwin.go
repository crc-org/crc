package preflight

import (
	"fmt"
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks(vmDriver string) {
	preflightCheckSucceedsOrFails(false,
		checkOcBinaryCached,
		"Checking if oc binary is cached",
		false,
	)

	switch vmDriver {
	case "virtualbox":
		preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckVirtualBoxInstalled.Name),
			checkVirtualBoxInstalled,
			"Checking if VirtualBox is Installed",
			config.GetBool(cmdConfig.WarnCheckVirtualBoxInstalled.Name),
		)
	}

	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckResolverFilePermissions.Name),
		checkResolverFilePermissions,
		fmt.Sprintf("Checking file permissions for %s", resolverFile),
		config.GetBool(cmdConfig.WarnCheckResolverFilePermissions.Name),
	)

	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckResolvConfFilePermissions.Name),
		checkResolvConfFilePermissions,
		fmt.Sprintf("Checking file permissions for %s", hostFile),
		config.GetBool(cmdConfig.WarnCheckResolvConfFilePermissions.Name),
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost(vmDriver string) {
	preflightCheckAndFix(false,
		checkOcBinaryCached,
		fixOcBinaryCached,
		"Caching oc binary",
		false,
	)

	switch vmDriver {
	case "virtualbox":
		preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckVirtualBoxInstalled.Name),
			checkVirtualBoxInstalled,
			fixVirtualBoxInstallation,
			"Setting up virtualization",
			config.GetBool(cmdConfig.WarnCheckVirtualBoxInstalled.Name),
		)
	}

	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckResolverFilePermissions.Name),
		checkResolverFilePermissions,
		fixResolverFilePermissions,
		fmt.Sprintf("Setting file permissions for %s", resolverFile),
		config.GetBool(cmdConfig.WarnCheckResolverFilePermissions.Name),
	)

	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckResolvConfFilePermissions.Name),
		checkResolvConfFilePermissions,
		fixResolvConfFilePermissions,
		fmt.Sprintf("Setting file permissions for %s", hostFile),
		config.GetBool(cmdConfig.WarnCheckResolvConfFilePermissions.Name),
	)
}
