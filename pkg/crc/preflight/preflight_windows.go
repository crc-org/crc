package preflight

import (
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
	preflightCheckSucceedsOrFails(config.GetBool(cmdConfig.SkipCheckBundleCached.Name),
		checkBundleCached,
		"Checking if CRC bundle is cached in '$HOME/.crc'",
		config.GetBool(cmdConfig.WarnCheckBundleCached.Name),
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
	preflightCheckAndFix(config.GetBool(cmdConfig.SkipCheckBundleCached.Name),
		checkBundleCached,
		fixBundleCached,
		"Unpacking bundle from the CRC binary",
		config.GetBool(cmdConfig.WarnCheckBundleCached.Name),
	)
}
