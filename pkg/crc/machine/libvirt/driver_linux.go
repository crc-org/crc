package libvirt

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/drivers/libvirt"
)

func CreateHost(machineConfig config.MachineConfig) *libvirt.Driver {
	libvirtDriver := libvirt.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, libvirtDriver.VMDriver)

	if machineConfig.NetworkMode == network.VSockMode {
		libvirtDriver.Network = "" // don't need to attach a network interface
		libvirtDriver.VSock = true
	} else {
		libvirtDriver.Network = DefaultNetwork
	}

	libvirtDriver.StoragePool = DefaultStoragePool
	return libvirtDriver
}
