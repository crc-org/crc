package config

import (
	"github.com/containers/common/pkg/strongunits"
	"github.com/crc-org/crc/v2/pkg/crc/network"
)

type MachineConfig struct {
	// CRC system bundle
	BundleName string

	// Virtual machine configuration
	Name string
	// Memory holds value in MiB
	Memory strongunits.MiB
	CPUs   uint
	// DiskSize holds value in GiB
	DiskSize          strongunits.GiB
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
