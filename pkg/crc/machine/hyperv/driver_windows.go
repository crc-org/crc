package hyperv

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/machine/drivers/hyperv"

	"github.com/code-ready/crc/pkg/crc/machine/config"

	winnet "github.com/code-ready/crc/pkg/os/windows/network"
)

func CreateHost(machineConfig config.MachineConfig) *hyperv.Driver {
	hypervDriver := hyperv.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	hypervDriver.CPU = machineConfig.CPUs
	hypervDriver.BundleName = machineConfig.BundleName

	// memory related settings
	hypervDriver.DisableDynamicMemory = true
	hypervDriver.Memory = machineConfig.Memory

	// Determine the Virtual Switch to be used
	_, switchName := winnet.SelectSwitchByNameOrDefault(AlternativeNetwork)
	hypervDriver.VirtualSwitch = switchName

	hypervDriver.ImageSourcePath = machineConfig.ImageSourcePath
	hypervDriver.ImageFormat = machineConfig.ImageFormat
	hypervDriver.SSHKeyPath = machineConfig.SSHKeyPath

	return hypervDriver
}
