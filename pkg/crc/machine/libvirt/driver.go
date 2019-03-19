package libvirt

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"

	"github.com/code-ready/crc/pkg/crc/machine/config"

	"github.com/code-ready/machine/libmachine/drivers"
)

type libvirtDriver struct {
	*drivers.BaseDriver

	// CRC System bundle
	BundlePath string

	// Driver specific configuration
	Memory    int
	CPU       int
	Network   string
	DiskPath  string
	CacheMode string
	IOMode    string
}

func CreateHost(machineConfig config.MachineConfig) *libvirtDriver {
	return &libvirtDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: machineConfig.Name,
			StorePath:   constants.MachineBaseDir,
			SSHUser:     constants.DefaultSSHUser,
		},
		BundlePath: machineConfig.BundlePath,
		Memory:     machineConfig.Memory,
		CPU:        machineConfig.CPUs,
		Network:    DefaultNetwork,
		// DiskPath should come from the bundle's metadata (unflattened)
		DiskPath:  filepath.Join(constants.MachineBaseDir, "machines", machineConfig.Name, "crc_libvirt_0.16.1", "crc"),
		CacheMode: DefaultCacheMode,
		IOMode:    DefaultIOMode,
	}
}
