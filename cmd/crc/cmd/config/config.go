package config

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/tray"
	"github.com/spf13/cast"
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
	HostNetworkAccess       = "host-network-access"
	HTTPProxy               = "http-proxy"
	HTTPSProxy              = "https-proxy"
	NoProxy                 = "no-proxy"
	ProxyCAFile             = "proxy-ca-file"
	ConsentTelemetry        = "consent-telemetry"
	EnableClusterMonitoring = "enable-cluster-monitoring"
	AutostartTray           = "autostart-tray"
)

func RegisterSettings(cfg *config.Config) {
	validateHostNetworkAccess := func(value interface{}) (bool, string) {
		mode := network.ParseMode(cfg.Get(NetworkMode).AsString())
		if mode != network.VSockMode {
			return false, fmt.Sprintf("%s can only be used with %s set to '%s'",
				HostNetworkAccess, NetworkMode, network.VSockMode)
		}
		return config.ValidateBool(value)
	}

	disableEnableTrayAutostart := func(key string, value interface{}) string {
		if cast.ToBool(value) {
			return fmt.Sprintf(
				"Successfully configured '%s' to '%s'. Run 'crc setup' for it to take effect.",
				key, cast.ToString(value),
			)
		}
		return fmt.Sprintf(
			"Successfully configured '%s' to '%s'. Run 'crc cleanup' and then 'crc setup' for it to take effect.",
			key, cast.ToString(value),
		)
	}

	// Start command settings in config
	cfg.AddSetting(Bundle, constants.DefaultBundlePath, config.ValidateBundlePath, config.SuccessfullyApplied,
		fmt.Sprintf("Bundle path (string, default '%s')", constants.DefaultBundlePath))
	cfg.AddSetting(CPUs, constants.DefaultCPUs, config.ValidateCPUs, config.RequiresRestartMsg,
		fmt.Sprintf("Number of CPU cores (must be greater than or equal to '%d')", constants.DefaultCPUs))
	cfg.AddSetting(Memory, constants.DefaultMemory, config.ValidateMemory, config.RequiresRestartMsg,
		fmt.Sprintf("Memory size in MiB (must be greater than or equal to '%d')", constants.DefaultMemory))
	cfg.AddSetting(DiskSize, constants.DefaultDiskSize, config.ValidateDiskSize, config.RequiresRestartMsg,
		fmt.Sprintf("Total size in GiB of the disk (must be greater than or equal to '%d')", constants.DefaultDiskSize))
	cfg.AddSetting(NameServer, "", config.ValidateIPAddress, config.SuccessfullyApplied,
		"IPv4 address of nameserver (string, like '1.1.1.1 or 8.8.8.8')")
	cfg.AddSetting(PullSecretFile, "", config.ValidatePath, config.SuccessfullyApplied,
		fmt.Sprintf("Path of image pull secret (download from %s)", constants.CrcLandingPageURL))
	cfg.AddSetting(DisableUpdateCheck, false, config.ValidateBool, config.SuccessfullyApplied,
		"Disable update check (true/false, default: false)")
	cfg.AddSetting(ExperimentalFeatures, false, config.ValidateBool, config.SuccessfullyApplied,
		"Enable experimental features (true/false, default: false)")
	cfg.AddSetting(NetworkMode, string(network.DefaultMode), network.ValidateMode, network.SuccessfullyAppliedMode,
		"Network mode (default or vsock)")
	cfg.AddSetting(HostNetworkAccess, false, validateHostNetworkAccess, config.SuccessfullyApplied,
		"Allow TCP/IP connections from the CodeReady Containers VM to services running on the host (true/false, default: false)")
	// System tray auto-start config
	cfg.AddSetting(AutostartTray, true, tray.ValidateTrayAutostart, disableEnableTrayAutostart,
		"Automatically start the tray (true/false, default: true)")
	// Proxy Configuration
	cfg.AddSetting(HTTPProxy, "", config.ValidateURI, config.SuccessfullyApplied,
		"HTTP proxy URL (string, like 'http://my-proxy.com:8443')")
	cfg.AddSetting(HTTPSProxy, "", config.ValidateURI, config.SuccessfullyApplied,
		"HTTPS proxy URL (string, like 'https://my-proxy.com:8443')")
	cfg.AddSetting(NoProxy, "", config.ValidateNoProxy, config.SuccessfullyApplied,
		"Hosts, ipv4 addresses or CIDR which do not use a proxy (string, comma-separated list such as '127.0.0.1,192.168.100.1/24')")
	cfg.AddSetting(ProxyCAFile, "", config.ValidatePath, config.SuccessfullyApplied,
		"Path to an HTTPS proxy certificate authority (CA)")

	cfg.AddSetting(EnableClusterMonitoring, false, config.ValidateBool, config.SuccessfullyApplied,
		"Enable cluster monitoring Operator (true/false, default: false)")

	// Telemeter Configuration
	cfg.AddSetting(ConsentTelemetry, "", config.ValidateYesNo, config.SuccessfullyApplied,
		"Consent to collection of anonymous usage data (yes/no)")
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
	settings := config.AllSettings()
	sort.Slice(settings, func(i, j int) bool {
		return less(settings[i].Name, settings[j].Name)
	})

	var buf bytes.Buffer
	writer := tabwriter.NewWriter(&buf, 0, 8, 1, ' ', tabwriter.TabIndent)
	for _, cfg := range settings {
		fmt.Fprintf(writer, "* %s\t%s\n", cfg.Name, cfg.Help)
	}
	writer.Flush()

	return buf.String()
}

func GetConfigCmd(config *config.Config) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config SUBCOMMAND [flags]",
		Short: "Modify crc configuration",
		Long: `Modifies crc configuration properties.
Properties: ` + "\n\n" + configurableFields(config),
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
