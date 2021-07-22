package config

import (
	"fmt"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/version"

	"github.com/spf13/cast"
)

type Preset string

const (
	Podman    Preset = "podman"
	Openshift Preset = "openshift"
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
	KubeAdminPassword       = "kubeadmin-password"
	PresetConfigurationKey  = "preset"
)

func RegisterSettings(cfg *Config) {
	validateTrayAutostart := func(value interface{}) (bool, string) {
		if runtime.GOOS == "linux" {
			return false, "Tray autostart is only supported on macOS and windows"
		}
		return ValidateBool(value)
	}

	validateHostNetworkAccess := func(value interface{}) (bool, string) {
		mode := GetNetworkMode(cfg)
		if mode != network.UserNetworkingMode {
			return false, fmt.Sprintf("%s can only be used with %s set to '%s'",
				HostNetworkAccess, NetworkMode, network.UserNetworkingMode)
		}
		return ValidateBool(value)
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
	cfg.AddSetting(Bundle, constants.DefaultBundlePath, ValidateBundlePath, SuccessfullyApplied,
		fmt.Sprintf("Bundle path (string, default '%s')", constants.DefaultBundlePath))
	cfg.AddSetting(CPUs, constants.DefaultCPUs, ValidateCPUs, RequiresRestartMsg,
		fmt.Sprintf("Number of CPU cores (must be greater than or equal to '%d')", constants.DefaultCPUs))
	cfg.AddSetting(Memory, constants.DefaultMemory, ValidateMemory, RequiresRestartMsg,
		fmt.Sprintf("Memory size in MiB (must be greater than or equal to '%d')", constants.DefaultMemory))
	cfg.AddSetting(DiskSize, constants.DefaultDiskSize, ValidateDiskSize, RequiresRestartMsg,
		fmt.Sprintf("Total size in GiB of the disk (must be greater than or equal to '%d')", constants.DefaultDiskSize))
	cfg.AddSetting(NameServer, "", ValidateIPAddress, SuccessfullyApplied,
		"IPv4 address of nameserver (string, like '1.1.1.1 or 8.8.8.8')")
	cfg.AddSetting(PullSecretFile, "", ValidatePath, SuccessfullyApplied,
		fmt.Sprintf("Path of image pull secret (download from %s)", constants.CrcLandingPageURL))
	cfg.AddSetting(DisableUpdateCheck, false, ValidateBool, SuccessfullyApplied,
		"Disable update check (true/false, default: false)")
	cfg.AddSetting(ExperimentalFeatures, false, ValidateBool, SuccessfullyApplied,
		"Enable experimental features (true/false, default: false)")

	if !version.IsInstaller() {
		cfg.AddSetting(NetworkMode, string(defaultNetworkMode()), network.ValidateMode, network.SuccessfullyAppliedMode,
			fmt.Sprintf("Network mode (%s or %s)", network.UserNetworkingMode, network.SystemNetworkingMode))
	}

	cfg.AddSetting(HostNetworkAccess, false, validateHostNetworkAccess, SuccessfullyApplied,
		"Allow TCP/IP connections from the CodeReady Containers VM to services running on the host (true/false, default: false)")
	// System tray auto-start config
	cfg.AddSetting(AutostartTray, true, validateTrayAutostart, disableEnableTrayAutostart,
		"Automatically start the tray (true/false, default: true)")
	// Proxy Configuration
	cfg.AddSetting(HTTPProxy, "", ValidateHTTPProxy, SuccessfullyApplied,
		"HTTP proxy URL (string, like 'http://my-proxy.com:8443')")
	cfg.AddSetting(HTTPSProxy, "", ValidateHTTPSProxy, SuccessfullyApplied,
		"HTTPS proxy URL (string, like 'https://my-proxy.com:8443')")
	cfg.AddSetting(NoProxy, "", ValidateNoProxy, SuccessfullyApplied,
		"Hosts, ipv4 addresses or CIDR which do not use a proxy (string, comma-separated list such as '127.0.0.1,192.168.100.1/24')")
	cfg.AddSetting(ProxyCAFile, "", ValidatePath, SuccessfullyApplied,
		"Path to an HTTPS proxy certificate authority (CA)")

	cfg.AddSetting(EnableClusterMonitoring, false, ValidateBool, SuccessfullyApplied,
		"Enable cluster monitoring Operator (true/false, default: false)")

	// Telemeter Configuration
	cfg.AddSetting(ConsentTelemetry, "", ValidateYesNo, SuccessfullyApplied,
		"Consent to collection of anonymous usage data (yes/no)")

	cfg.AddSetting(KubeAdminPassword, "", ValidateString, SuccessfullyApplied,
		"User defined kubeadmin password")

	cfg.AddSetting(PresetConfigurationKey, string(Openshift), validatePreset, RequiresDeleteMsg,
		fmt.Sprintf("Virtal machine preset (alpha feature - valid values are: %s or %s)", Podman, Openshift))
}

func defaultNetworkMode() network.Mode {
	if version.IsInstaller() {
		return network.UserNetworkingMode
	}
	return network.SystemNetworkingMode
}

func GetNetworkMode(config Storage) network.Mode {
	if version.IsInstaller() {
		return network.UserNetworkingMode
	}
	return network.ParseMode(config.Get(NetworkMode).AsString())
}

func validatePreset(i interface{}) (bool, string) {
	switch Preset(cast.ToString(i)) {
	case Podman, Openshift:
		return true, ""
	default:
		return false, fmt.Sprintf("Unknown preset. Only %s and %s are valid.", Podman, Openshift)
	}
}
