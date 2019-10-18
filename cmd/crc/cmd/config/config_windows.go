package config

import (
	cfg "github.com/code-ready/crc/pkg/crc/config"
)

var (
	// Preflight checks
	SkipCheckAdministratorUser   = cfg.AddSetting("skip-check-administrator-user", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckAdministratorUser   = cfg.AddSetting("warn-check-administrator-user", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	SkipCheckWindowsVersionCheck = cfg.AddSetting("skip-check-windows-version", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckWindowsVersionCheck = cfg.AddSetting("warn-check-windows-version", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	SkipCheckHyperVInstalled     = cfg.AddSetting("skip-check-hyperv-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckHyperVInstalled     = cfg.AddSetting("warn-check-hyperv-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	SkipCheckUserInHyperVGroup   = cfg.AddSetting("skip-check-user-in-hyperv-group", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckUserInHyperVGroup   = cfg.AddSetting("warn-check-user-in-hyperv-group", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	SkipCheckHyperVSwitch        = cfg.AddSetting("skip-check-hyperv-switch", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
	WarnCheckHyperVSwitch        = cfg.AddSetting("warn-check-hyperv-switch", nil, []cfg.ValidationFnType{cfg.ValidateBool}, []cfg.SetFn{cfg.SuccessfullyApplied})
)
