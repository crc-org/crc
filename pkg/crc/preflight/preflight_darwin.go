package preflight

import (
	"fmt"
	cmdConfig "github.com/code-ready/crc/cmd/crc/cmd/config"
	"github.com/code-ready/crc/pkg/crc/config"
)

// StartPreflightChecks performs the preflight checks before starting the cluster
func StartPreflightChecks(vmDriver string) {
	preflightCheckSucceedsOrFails(false,
		checkIfRunningAsNormalUser,
		"Checking if running as root",
		false,
	)
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
	case "hyperkit":
		preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckHyperKitInstalled.Name),
			checkHyperKitInstalled,
			"Checking if HyperKit is installed",
			config.GetBool(cmdConfig.WarnCheckHyperKitInstalled.Name),
		)
		preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckHyperKitDriver.Name),
			checkMachineDriverHyperKitInstalled,
			"Checking if crc-driver-hyperkit is installed",
			config.GetBool(cmdConfig.WarnCheckHyperKitDriver.Name),
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
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckBundleCached.Name),
		checkBundleCached,
		"Checking if CRC bundle is cached in '$HOME/.crc'",
		config.GetBool(cmdConfig.WarnCheckBundleCached.Name),
	)
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func SetupHost(vmDriver string) {
	preflightCheckAndFix(false,
		checkIfRunningAsNormalUser,
		fixRunAsNormalUser,
		"Checking if running as root",
		false,
	)
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
			"Setting up virtualization with VirtualBox",
			config.GetBool(cmdConfig.WarnCheckVirtualBoxInstalled.Name),
		)
	case "hyperkit":
		preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckHyperKitInstalled.Name),
			checkHyperKitInstalled,
			fixHyperKitInstallation,
			"Setting up virtualization with HyperKit",
			config.GetBool(cmdConfig.WarnCheckHyperKitInstalled.Name),
		)
		preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckHyperKitDriver.Name),
			checkMachineDriverHyperKitInstalled,
			fixMachineDriverHyperKitInstalled,
			"Installing crc-machine-hyperkit",
			config.GetBool(cmdConfig.WarnCheckHyperKitDriver.Name),
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
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckBundleCached.Name),
		checkBundleCached,
		fixBundleCached,
		"Unpacking bundle from the CRC binary",
		config.GetBool(cmdConfig.WarnCheckBundleCached.Name),
	)
}
