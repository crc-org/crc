package config

import (
	cfg "github.com/code-ready/crc/pkg/crc/config"
)

var (
	// Preflight checks
	SkipCheckResolverFilePermissions = cfg.AddSetting("skip-check-resolver-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckResolverFilePermissions = cfg.AddSetting("warn-check-resolver-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	SkipCheckHostsFilePermissions    = cfg.AddSetting("skip-check-hosts-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckHostsFilePermissions    = cfg.AddSetting("warn-check-hosts-file-permissions", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	SkipCheckHyperKitDriver          = cfg.AddSetting("skip-check-hyperkit-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckHyperKitDriver          = cfg.AddSetting("warn-check-hyperkit-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	SkipCheckHyperKitInstalled       = cfg.AddSetting("skip-check-hyperkit-driver", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckHyperKitInstalled       = cfg.AddSetting("warn-check-hyperkit-driver", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
)
