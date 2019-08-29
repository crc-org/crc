package config

import (
	"sort"
	"strings"

	cfg "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/spf13/cobra"
)

var (
	// Start command settings in config
	VMDriver       = cfg.AddSetting("vm-driver", machine.DefaultDriver.Driver, []cfg.ValidationFnType{cfg.ValidateDriver})
	Bundle         = cfg.AddSetting("bundle", nil, []cfg.ValidationFnType{cfg.ValidateBundle})
	CPUs           = cfg.AddSetting("cpus", constants.DefaultCPUs, []cfg.ValidationFnType{cfg.ValidateCPUs})
	Memory         = cfg.AddSetting("memory", constants.DefaultMemory, []cfg.ValidationFnType{cfg.ValidateMemory})
	NameServer     = cfg.AddSetting("nameserver", nil, []cfg.ValidationFnType{cfg.ValidateIpAddress})
	PullSecretFile = cfg.AddSetting("pull-secret-file", nil, []cfg.ValidationFnType{cfg.ValidatePath})

	// Preflight checks
	SkipCheckVirtualBoxInstalled = cfg.AddSetting("skip-check-virtualbox-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckVirtualBoxInstalled = cfg.AddSetting("warn-check-virtualbox-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckBundleCached        = cfg.AddSetting("skip-check-bundle-cached", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckBundleCached        = cfg.AddSetting("warn-check-bundle-cached", true, []cfg.ValidationFnType{cfg.ValidateBool})
)

var (
	ConfigCmd = &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modifies crc configuration properties.",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
)

func init() {
	ConfigCmd.Long = `Modifies crc configuration properties. Some of the configuration properties are equivalent
to the options that you set when you run the 'crc start' command.
Configurable properties (enter as SUBCOMMAND): ` + "\n\n" + configurableFields()
}

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
			} else {
				// ignore skip-/warn- prefix
				return lhsKey[4:] < rhsKey[4:]
			}
		} else {
			// lhs is preflight, rhs is not preflight
			return false
		}
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
	keys := cfg.AllConfigKeys()
	sort.Slice(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})
	for _, key := range keys {
		fields = append(fields, " * "+key)
	}
	return strings.Join(fields, "\n")
}
