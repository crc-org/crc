package services

import (
	"github.com/code-ready/crc/pkg/crc/machine/bundle"

	"github.com/code-ready/machine/libmachine/drivers"
)

type ServicePreStartConfig struct {
	Name           string
	BundleMetadata bundle.CrcBundleInfo
}

type ServicePreStartResult struct {
	Name    string
	Success bool
	Error   string
}

type ServicePostStartConfig struct {
	Name           string
	Driver         drivers.Driver
	BundleMetadata bundle.CrcBundleInfo
	IP             string
	HostIP         string
}

type ServicePostStartResult struct {
	Name    string
	Success bool
	Error   string
}
