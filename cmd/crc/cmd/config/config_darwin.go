package config

import (
	cfg "github.com/code-ready/crc/pkg/crc/config"
)

var (
	// Preflight checks
	SkipCheckResolverFilePermissions   = cfg.AddSetting("skip-check-resolver-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckResolverFilePermissions   = cfg.AddSetting("warn-check-resolver-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckResolvConfFilePermissions = cfg.AddSetting("skip-check-resolv-conf-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckResolvConfFilePermissions = cfg.AddSetting("warn-check-resolv-conf-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckHyperKitDriver            = cfg.AddSetting("skip-check-hyperkit-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckHyperKitDriver            = cfg.AddSetting("warn-check-hyperkit-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckHyperKitInstalled         = cfg.AddSetting("skip-check-hyperkit-driver", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckHyperKitInstalled         = cfg.AddSetting("warn-check-hyperkit-driver", nil, []cfg.ValidationFnType{cfg.ValidateBool})
)
