package libvirt

import (
	"path/filepath"

	"github.com/code-ready/machine/drivers/libvirt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/machine/libmachine/drivers"
)

func CreateHost(machineConfig config.MachineConfig, baseDriver *drivers.BaseDriver) drivers.Driver {
	return &libvirt.Driver{
		BaseDriver: baseDriver,
		Memory:     machineConfig.Memory,
		CPU:        machineConfig.CPUs,
		Network:    DefaultNetwork,
		CacheMode:  DefaultCacheMode,
		IOMode:     DefaultIOMode,
		// This force to add entry of DiskPath under crc machine config.json
		DiskPath:    filepath.Join(constants.MachineBaseDir, "machines", machineConfig.Name, constants.DefaultName),
		DiskPathURL: machineConfig.DiskPathURL,
		SSHKeyPath:  machineConfig.SSHKeyPath,
	}
}
