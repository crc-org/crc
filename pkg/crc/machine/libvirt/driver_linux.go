package libvirt

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/crc-org/machine/drivers/libvirt"
	"github.com/crc-org/machine/libmachine/drivers"
)

func CreateHost(machineConfig config.MachineConfig) *libvirt.Driver {
	libvirtDriver := libvirt.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, libvirtDriver.VMDriver)

	if machineConfig.NetworkMode == network.UserNetworkingMode {
		libvirtDriver.Network = "" // don't need to attach a network interface
		libvirtDriver.VSock = true
	} else {
		libvirtDriver.Network = DefaultNetwork
	}

	libvirtDriver.StoragePool = DefaultStoragePool
	libvirtDriver.SharedDirs = configureShareDirs(machineConfig)

	return libvirtDriver
}

func configureShareDirs(machineConfig config.MachineConfig) []drivers.SharedDir {
	var sharedDirs []drivers.SharedDir
	for i, dir := range machineConfig.SharedDirs {
		sharedDir := drivers.SharedDir{
			Source: dir,
			Target: dir,
			Tag:    fmt.Sprintf("dir%d", i),
			Type:   "virtiofs",
		}
		sharedDirs = append(sharedDirs, sharedDir)
	}
	return sharedDirs
}
