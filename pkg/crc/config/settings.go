package config

import (
	"fmt"
	"runtime"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/version"
)

const (
	Bundle                   = "bundle"
	CPUs                     = "cpus"
	Memory                   = "memory"
	DiskSize                 = "disk-size"
	NameServer               = "nameserver"
	PullSecretFile           = "pull-secret-file"
	DisableUpdateCheck       = "disable-update-check"
	ExperimentalFeatures     = "enable-experimental-features"
	NetworkMode              = "network-mode"
	HostNetworkAccess        = "host-network-access"
	HTTPProxy                = "http-proxy"
	HTTPSProxy               = "https-proxy"
	NoProxy                  = "no-proxy"
	ProxyCAFile              = "proxy-ca-file"
	ConsentTelemetry         = "consent-telemetry"
	EnableClusterMonitoring  = "enable-cluster-monitoring"
	KubeAdminPassword        = "kubeadmin-password"
	Preset                   = "preset"
	EnableSharedDirs         = "enable-shared-dirs"
	SharedDirPassword        = "shared-dir-password" // #nosec G101
	IngressHTTPPort          = "ingress-http-port"
	IngressHTTPSPort         = "ingress-https-port"
	EmergencyLogin           = "enable-emergency-login"
	PersistentVolumeSize     = "persistent-volume-size"
	EnableBundleQuayFallback = "enable-bundle-quay-fallback"
)

func RegisterSettings(cfg *Config) {
	validateHostNetworkAccess := func(value interface{}) (bool, string) {
		mode := GetNetworkMode(cfg)
		if mode != network.UserNetworkingMode {
			return false, fmt.Sprintf("%s can only be used with %s set to '%s'",
				HostNetworkAccess, NetworkMode, network.UserNetworkingMode)
		}
		return ValidateBool(value)
	}

	validCPUs := func(value interface{}) (bool, string) {
		return validateCPUs(value, GetPreset(cfg))
	}

	validMemory := func(value interface{}) (bool, string) {
		return validateMemory(value, GetPreset(cfg))
	}

	validBundlePath := func(value interface{}) (bool, string) {
		return validateBundlePath(value, GetPreset(cfg))
	}

	validateSmbSharedDirs := func(value interface{}) (bool, string) {
		if !cfg.Get(HostNetworkAccess).AsBool() {
			return false, fmt.Sprintf("%s can only be used with %s set to 'true'",
				EnableSharedDirs, HostNetworkAccess)
		}
		if cfg.Get(SharedDirPassword).IsDefault {
			return false, fmt.Sprintf("Please set '%s' first to enable shared directories", SharedDirPassword)
		}
		return ValidateBool(value)
	}

	// Preset setting should be on top because CPUs/Memory config depend on it.
	cfg.AddSetting(Preset, version.GetDefaultPreset().String(), validatePreset, RequiresDeleteAndSetupMsg,
		fmt.Sprintf("Virtual machine preset (valid values are: %s)", preset.AllPresets()))
	// Start command settings in config
	cfg.AddSetting(Bundle, defaultBundlePath(cfg), validBundlePath, SuccessfullyApplied, BundleHelpMsg(cfg))
	cfg.AddSetting(CPUs, defaultCPUs(cfg), validCPUs, RequiresRestartMsg,
		fmt.Sprintf("Number of CPU cores (must be greater than or equal to '%d')", defaultCPUs(cfg)))
	cfg.AddSetting(Memory, defaultMemory(cfg), validMemory, RequiresRestartMsg,
		fmt.Sprintf("Memory size in MiB (must be greater than or equal to '%d')", defaultMemory(cfg)))
	cfg.AddSetting(DiskSize, constants.DefaultDiskSize, validateDiskSize, RequiresRestartMsg,
		fmt.Sprintf("Total size in GiB of the disk (must be greater than or equal to '%d')", constants.DefaultDiskSize))
	cfg.AddSetting(NameServer, "", validateIPAddress, SuccessfullyApplied,
		"IPv4 address of nameserver (string, like '1.1.1.1 or 8.8.8.8')")
	cfg.AddSetting(PullSecretFile, "", validatePath, SuccessfullyApplied,
		fmt.Sprintf("Path of image pull secret (download from %s)", constants.CrcLandingPageURL))
	cfg.AddSetting(DisableUpdateCheck, false, ValidateBool, SuccessfullyApplied,
		"Disable update check (true/false, default: false)")
	cfg.AddSetting(ExperimentalFeatures, false, ValidateBool, SuccessfullyApplied,
		"Enable experimental features (true/false, default: false)")
	cfg.AddSetting(EmergencyLogin, false, ValidateBool, SuccessfullyApplied,
		"Enable emergency login for 'core' user. Password is randomly generated. (true/false, default: false)")
	cfg.AddSetting(PersistentVolumeSize, constants.DefaultPersistentVolumeSize, validatePersistentVolumeSize, SuccessfullyApplied,
		fmt.Sprintf("Total size in GiB of the persistent volume used by the CSI driver for %s preset (must be greater than or equal to '%d')", preset.Microshift, constants.DefaultPersistentVolumeSize))

	// Shared directories configs
	if runtime.GOOS == "windows" {
		cfg.AddSetting(SharedDirPassword, Secret(""), validateString, SuccessfullyApplied,
			"Password used while using CIFS/SMB file sharing (It is the password for the current logged in user)")

		cfg.AddSetting(EnableSharedDirs, false, validateSmbSharedDirs, SuccessfullyApplied,
			"Mounts host's user profile folder at '/' in the CRC VM (true/false, default: false)")
	} else {
		cfg.AddSetting(EnableSharedDirs, true, ValidateBool, SuccessfullyApplied,
			"Mounts host's home directory at '/' in the CRC VM (true/false, default: true)")
	}

	if !version.IsInstaller() {
		cfg.AddSetting(NetworkMode, string(defaultNetworkMode()), network.ValidateMode, network.SuccessfullyAppliedMode,
			fmt.Sprintf("Network mode (%s or %s)", network.UserNetworkingMode, network.SystemNetworkingMode))
	}

	cfg.AddSetting(HostNetworkAccess, false, validateHostNetworkAccess, RequiresCleanupAndSetupMsg,
		"Allow TCP/IP connections from the CRC VM to services running on the host (true/false, default: false)")
	// Proxy Configuration
	cfg.AddSetting(HTTPProxy, "", validateHTTPProxy, SuccessfullyApplied,
		"HTTP proxy URL (string, like 'http://my-proxy.com:8443')")
	cfg.AddSetting(HTTPSProxy, "", validateHTTPSProxy, SuccessfullyApplied,
		"HTTPS proxy URL (string, like 'https://my-proxy.com:8443')")
	cfg.AddSetting(NoProxy, "", validateNoProxy, SuccessfullyApplied,
		"Hosts, ipv4 addresses or CIDR which do not use a proxy (string, comma-separated list such as '127.0.0.1,192.168.100.1/24')")
	cfg.AddSetting(ProxyCAFile, Path(""), validatePath, SuccessfullyApplied,
		"Path to an HTTPS proxy certificate authority (CA)")

	cfg.AddSetting(EnableClusterMonitoring, false, ValidateBool, SuccessfullyApplied,
		"Enable cluster monitoring Operator (true/false, default: false)")

	// Telemeter Configuration
	cfg.AddSetting(ConsentTelemetry, "", validateYesNo, SuccessfullyApplied,
		"Consent to collection of anonymous usage data (yes/no)")

	cfg.AddSetting(KubeAdminPassword, "", validateString, SuccessfullyApplied,
		"User defined kubeadmin password")
	cfg.AddSetting(IngressHTTPPort, constants.OpenShiftIngressHTTPPort, validatePort, RequiresHTTPPortChangeWarning,
		fmt.Sprintf("HTTP port to use for OpenShift ingress/routes on the host (1024-65535, default: %d)", constants.OpenShiftIngressHTTPPort))
	cfg.AddSetting(IngressHTTPSPort, constants.OpenShiftIngressHTTPSPort, validatePort, RequiresHTTPSPortChangeWarning,
		fmt.Sprintf("HTTPS port to use for OpenShift ingress/routes on the host (1024-65535, default: %d)", constants.OpenShiftIngressHTTPSPort))

	cfg.AddSetting(EnableBundleQuayFallback, false, ValidateBool, SuccessfullyApplied,
		"If bundle download from the default location fails, fallback to quay.io (true/false, default: false)")

	if err := cfg.RegisterNotifier(Preset, presetChanged); err != nil {
		logging.Debugf("Failed to register notifier for Preset: %v", err)
	}
}

func presetChanged(cfg *Config, _ string, _ interface{}) {
	UpdateDefaults(cfg)
}

func defaultCPUs(cfg Storage) int {
	return constants.GetDefaultCPUs(GetPreset(cfg))
}

func defaultMemory(cfg Storage) int {
	return constants.GetDefaultMemory(GetPreset(cfg))
}

func defaultBundlePath(cfg Storage) string {
	return constants.GetDefaultBundlePath(GetPreset(cfg))
}

func GetPreset(config Storage) preset.Preset {
	return preset.ParsePreset(config.Get(Preset).AsString())
}

func defaultNetworkMode() network.Mode {
	if runtime.GOOS != "linux" {
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

func revalidateSettingsValue(cfg *Config, key string) error {
	if err := cfg.validate(key, cfg.Get(key).Value); err != nil {
		logging.Debugf("'%s' value is invalid: %v", key, err)
		if _, err := cfg.Unset(key); err != nil {
			return err
		}
		logging.Warnf("Value for config setting '%s' is invalid, resetting it to its default value", key)
	}

	return nil
}

func UpdateDefaults(cfg *Config) {
	RegisterSettings(cfg)

	// The `memory` and `cpus` values are preset-dependent.
	// When they are lower than the preset requirements, this code
	// automatically resets them to their default value.
	if err := revalidateSettingsValue(cfg, Memory); err != nil {
		logging.Infof("failed to validate Memory setting: %v", err)
	}
	if err := revalidateSettingsValue(cfg, CPUs); err != nil {
		logging.Infof("failed to validate CPUs setting: %v", err)
	}
}

func BundleHelpMsg(cfg *Config) string {
	return fmt.Sprintf("Bundle path/URI - absolute or local path, http, https or docker URI (string, like 'https://foo.com/%s', 'docker://quay.io/myorg/%s:%s' default '%s' )",
		constants.GetDefaultBundle(GetPreset(cfg)), constants.GetDefaultBundle(GetPreset(cfg)), version.GetCRCVersion(), defaultBundlePath(cfg))
}
