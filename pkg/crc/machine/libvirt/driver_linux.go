package libvirt

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/machine/drivers/libvirt"
)

func CreateHost(machineConfig config.MachineConfig) *libvirt.Driver {
	libvirtDriver := libvirt.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	libvirtDriver.CPU = machineConfig.CPUs
	libvirtDriver.Memory = machineConfig.Memory
	libvirtDriver.BundleName = machineConfig.BundleName
	libvirtDriver.Network = DefaultNetwork
	libvirtDriver.ImageSourcePath = machineConfig.ImageSourcePath
	libvirtDriver.ImageFormat = machineConfig.ImageFormat
	libvirtDriver.SSHKeyPath = machineConfig.SSHKeyPath

	return libvirtDriver
}
