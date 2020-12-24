package hyperv

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/drivers/hyperv"
	winnet "github.com/code-ready/crc/pkg/os/windows/network"
)

func CreateHost(machineConfig config.MachineConfig) *hyperv.Driver {
	hypervDriver := hyperv.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, hypervDriver.VMDriver)

	hypervDriver.DisableDynamicMemory = true

	if machineConfig.NetworkMode == network.VSockMode {

		hypervDriver.VirtualSwitch = ""
	} else {
		// Determine the Virtual Switch to be used
		_, switchName := winnet.SelectSwitchByNameOrDefault(AlternativeNetwork)
		hypervDriver.VirtualSwitch = switchName
	}

	return hypervDriver
}
