package config

import "github.com/code-ready/crc/pkg/crc/network"

type MachineConfig struct {
	// CRC system bundle
	BundleName string

	// Virtual machine configuration
	Name              string
	Memory            int
	CPUs              int
	DiskSize          int
	ImageSourcePath   string
	ImageFormat       string
	SSHKeyPath        string
	KubeConfig        string
	SharedDirs        []string
	SharedDirPassword string
	SharedDirUsername string

	// macOS specific configuration
	KernelCmdLine string
	Initramfs     string
	Kernel        string

	// Experimental features
	NetworkMode network.Mode
}
