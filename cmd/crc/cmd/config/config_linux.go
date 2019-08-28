package config

import (
	cfg "github.com/code-ready/crc/pkg/crc/config"
)

var (
	// Preflight checks
	SkipCheckVirtEnabled             = cfg.AddSetting("skip-check-virt-enabled", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckVirtEnabled             = cfg.AddSetting("warn-check-virt-enabled", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckKvmEnabled              = cfg.AddSetting("skip-check-kvm-enabled", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckKvmEnabled              = cfg.AddSetting("warn-check-kvm-enabled", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckLibvirtInstalled        = cfg.AddSetting("skip-check-libvirt-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckLibvirtInstalled        = cfg.AddSetting("warn-check-libvirt-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckLibvirtEnabled          = cfg.AddSetting("skip-check-libvirt-enabled", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckLibvirtEnabled          = cfg.AddSetting("warn-check-libvirt-enabled", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckLibvirtRunning          = cfg.AddSetting("skip-check-libvirt-running", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckLibvirtRunning          = cfg.AddSetting("warn-check-libvirt-running", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckUserInLibvirtGroup      = cfg.AddSetting("skip-check-user-in-libvirt-group", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckUserInLibvirtGroup      = cfg.AddSetting("warn-check-user-in-libvirt-group", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckIPForwarding            = cfg.AddSetting("skip-check-ip-forwarding", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckIPForwarding            = cfg.AddSetting("warn-check-ip-forwarding", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckLibvirtDriver           = cfg.AddSetting("skip-check-libvirt-driver", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckLibvirtDriver           = cfg.AddSetting("warn-check-libvirt-driver", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckCrcNetwork              = cfg.AddSetting("skip-check-crc-network", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckCrcNetwork              = cfg.AddSetting("warn-check-crc-network", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckCrcNetworkActive        = cfg.AddSetting("skip-check-crc-network-active", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckCrcNetworkActive        = cfg.AddSetting("warn-check-crc-network-active", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckCrcDnsmasqFile          = cfg.AddSetting("skip-check-crc-dnsmasq-file", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckCrcDnsmasqFile          = cfg.AddSetting("warn-check-crc-dnsmasq-file", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckCrcNetworkManagerConfig = cfg.AddSetting("skip-check-network-manager-config", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckCrcNetworkManagerConfig = cfg.AddSetting("warn-check-network-manager-config", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckNetworkManagerInstalled = cfg.AddSetting("warn-check-network-manager-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckNetworkManagerInstalled = cfg.AddSetting("skip-check-network-manager-installed", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	WarnCheckNetworkManagerRunning   = cfg.AddSetting("warn-check-network-manager-running", nil, []cfg.ValidationFnType{cfg.ValidateBool})
	SkipCheckNetworkManagerRunning   = cfg.AddSetting("skip-check-network-manager-running", nil, []cfg.ValidationFnType{cfg.ValidateBool})
)
