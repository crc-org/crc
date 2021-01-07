package config

import (
	"sort"
	"strings"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/spf13/cobra"
)

const (
	Bundle                  = "bundle"
	CPUs                    = "cpus"
	Memory                  = "memory"
	DiskSize                = "disk-size"
	NameServer              = "nameserver"
	PullSecretFile          = "pull-secret-file"
	DisableUpdateCheck      = "disable-update-check"
	ExperimentalFeatures    = "enable-experimental-features"
	NetworkMode             = "network-mode"
	HTTPProxy               = "http-proxy"
	HTTPSProxy              = "https-proxy"
	NoProxy                 = "no-proxy"
	ProxyCAFile             = "proxy-ca-file"
	ConsentTelemetry        = "consent-telemetry"
	EnableClusterMonitoring = "enable-cluster-monitoring"
)

func RegisterSettings(cfg *config.Config) {
	// Start command settings in config
	cfg.AddSetting(Bundle, constants.DefaultBundlePath, config.ValidateBundle, config.SuccessfullyApplied)
	cfg.AddSetting(CPUs, constants.DefaultCPUs, config.ValidateCPUs, config.RequiresRestartMsg)
	cfg.AddSetting(Memory, constants.DefaultMemory, config.ValidateMemory, config.RequiresRestartMsg)
	cfg.AddSetting(DiskSize, constants.DefaultDiskSize, config.ValidateDiskSize, config.RequiresRestartMsg)
	cfg.AddSetting(NameServer, "", config.ValidateIPAddress, config.SuccessfullyApplied)
	cfg.AddSetting(PullSecretFile, "", config.ValidatePath, config.SuccessfullyApplied)
	cfg.AddSetting(DisableUpdateCheck, false, config.ValidateBool, config.SuccessfullyApplied)
	cfg.AddSetting(ExperimentalFeatures, false, config.ValidateBool, config.SuccessfullyApplied)
	cfg.AddSetting(NetworkMode, string(network.DefaultMode), network.ValidateMode, network.SuccessfullyAppliedMode)
	// Proxy Configuration
	cfg.AddSetting(HTTPProxy, "", config.ValidateURI, config.SuccessfullyApplied)
	cfg.AddSetting(HTTPSProxy, "", config.ValidateURI, config.SuccessfullyApplied)
	cfg.AddSetting(NoProxy, "", config.ValidateNoProxy, config.SuccessfullyApplied)
	cfg.AddSetting(ProxyCAFile, "", config.ValidatePath, config.SuccessfullyApplied)

	cfg.AddSetting(EnableClusterMonitoring, false, config.ValidateBool, config.SuccessfullyApplied)

	// Telemeter Configuration
	cfg.AddSetting(ConsentTelemetry, "", config.ValidateYesNo, config.SuccessfullyApplied)
}

func isPreflightKey(key string) bool {
	return strings.HasPrefix(key, "skip-")
}

// less is used to sort the config keys. We want to sort first the regular keys, and
// then the keys related to preflight starting with a skip- prefix.
func less(lhsKey, rhsKey string) bool {
	if isPreflightKey(lhsKey) {
		if isPreflightKey(rhsKey) {
			// ignore skip prefix
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

func configurableFields(config *config.Config) string {
	var fields []string
	var keys []string

	for key := range config.AllConfigs() {
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

func GetConfigCmd(config *config.Config) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modify crc configuration",
		Long: `Modifies crc configuration properties.
Configurable properties (enter as SUBCOMMAND): ` + "\n\n" + configurableFields(config),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	configCmd.AddCommand(configGetCmd(config))
	configCmd.AddCommand(configSetCmd(config))
	configCmd.AddCommand(configUnsetCmd(config))
	configCmd.AddCommand(configViewCmd(config))
	return configCmd
}
