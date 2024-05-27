package config

import "github.com/crc-org/crc/v2/pkg/crc/network"

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

	// Experimental features
	NetworkMode network.Mode
}
