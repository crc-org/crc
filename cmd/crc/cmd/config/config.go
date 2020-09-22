package config

import (
	"sort"
	"strings"

	cfg "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/spf13/cobra"
)

const (
	Bundle               = "bundle"
	CPUs                 = "cpus"
	Memory               = "memory"
	NameServer           = "nameserver"
	PullSecretFile       = "pull-secret-file"
	DisableUpdateCheck   = "disable-update-check"
	ExperimentalFeatures = "enable-experimental-features"
	HTTPProxy            = "http-proxy"
	HTTPSProxy           = "https-proxy"
	NoProxy              = "no-proxy"
	ProxyCAFile          = "proxy-ca-file"
)

func RegisterSettings() {
	// Start command settings in config
	cfg.AddSetting(Bundle, constants.DefaultBundlePath, cfg.ValidateBundle, cfg.SuccessfullyApplied)
	cfg.AddSetting(CPUs, constants.DefaultCPUs, cfg.ValidateCPUs, cfg.RequiresRestartMsg)
	cfg.AddSetting(Memory, constants.DefaultMemory, cfg.ValidateMemory, cfg.RequiresRestartMsg)
	cfg.AddSetting(NameServer, "", cfg.ValidateIPAddress, cfg.SuccessfullyApplied)
	cfg.AddSetting(PullSecretFile, "", cfg.ValidatePath, cfg.SuccessfullyApplied)
	cfg.AddSetting(DisableUpdateCheck, false, cfg.ValidateBool, cfg.SuccessfullyApplied)
	// Proxy Configuration
	cfg.AddSetting(ExperimentalFeatures, false, cfg.ValidateBool, cfg.SuccessfullyApplied)
	cfg.AddSetting(HTTPProxy, "", cfg.ValidateURI, cfg.SuccessfullyApplied)
	cfg.AddSetting(HTTPSProxy, "", cfg.ValidateURI, cfg.SuccessfullyApplied)
	cfg.AddSetting(NoProxy, "", cfg.ValidateNoProxy, cfg.SuccessfullyApplied)
	cfg.AddSetting(ProxyCAFile, "", cfg.ValidatePath, cfg.SuccessfullyApplied)
}

var (
	configCmd = &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modify crc configuration",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
)

func isPreflightKey(key string) bool {
	return strings.HasPrefix(key, "skip-") || strings.HasPrefix(key, "warn-")
}

// less is used to sort the config keys. We want to sort first the regular keys, and
// then the keys related to preflight starting with a skip- or warn- prefix. We want
// these preflight keys to be grouped by pair: 'skip-bar', 'warn-bar', 'skip-foo', 'warn-foo'
// would be sorted in that order.
func less(lhsKey, rhsKey string) bool {
	if isPreflightKey(lhsKey) {
		if isPreflightKey(rhsKey) {
			// lhs is preflight, rhs is preflight
			if lhsKey[4:] == rhsKey[4:] {
				// we want skip-foo before warn-foo
				return lhsKey < rhsKey
			}
			// ignore skip-/warn- prefix
			return lhsKey[4:] < rhsKey[4:]
		}
		// lhs is preflight, rhs is not preflight
		return false
	}

	if isPreflightKey(rhsKey) {
		// lhs is not preflight, rhs is preflight
		return true
	}

	// lhs is not preflight, rhs is not preflight
	return lhsKey < rhsKey
}

func configurableFields() string {
	var fields []string
	var keys []string

	for key := range cfg.AllConfigs() {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})
	for _, key := range keys {
		fields = append(fields, " * "+key)
	}
	return strings.Join(fields, "\n")
}

func GetConfigCmd() *cobra.Command {
	/* Delay generation of configCmd.Long as much as possible as some parts of crc may have registered more
	 * fields after init() time but before the command is registered
	 */
	configCmd.Long = `Modifies crc configuration properties.
Configurable properties (enter as SUBCOMMAND): ` + "\n\n" + configurableFields()

	return configCmd
}
