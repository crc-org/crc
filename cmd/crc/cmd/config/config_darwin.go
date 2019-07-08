package config

import (
	validations "github.com/code-ready/crc/pkg/crc/config"
)

var (
	// Preflight checks
	SkipCheckResolverFilePermissions   = createSetting("skip-check-resolver-file-permissions", nil, []validationFnType{validations.ValidateBool})
	WarnCheckResolverFilePermissions   = createSetting("warn-check-resolver-file-permissions", nil, []validationFnType{validations.ValidateBool})
	SkipCheckResolvConfFilePermissions = createSetting("skip-check-resolv-conf-file-permissions", nil, []validationFnType{validations.ValidateBool})
	WarnCheckResolvConfFilePermissions = createSetting("warn-check-resolv-conf-file-permissions", nil, []validationFnType{validations.ValidateBool})
	SkipCheckHyperKitDriver            = createSetting("skip-check-hyperkit-installed", nil, []validationFnType{validations.ValidateBool})
	WarnCheckHyperKitDriver            = createSetting("warn-check-hyperkit-installed", nil, []validationFnType{validations.ValidateBool})
	SkipCheckHyperKitInstalled         = createSetting("skip-check-hyperkit-driver", nil, []validationFnType{validations.ValidateBool})
	WarnCheckHyperKitInstalled         = createSetting("warn-check-hyperkit-driver", nil, []validationFnType{validations.ValidateBool})
)
