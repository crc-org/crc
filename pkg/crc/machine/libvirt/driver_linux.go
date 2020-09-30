package libvirt

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/machine/drivers/libvirt"
)

func CreateHost(machineConfig config.MachineConfig) *libvirt.Driver {
	libvirtDriver := libvirt.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, libvirtDriver.VMDriver)

	libvirtDriver.Network = DefaultNetwork
	libvirtDriver.StoragePool = DefaultStoragePool

	return libvirtDriver
}
