package hyperv

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/machine/drivers/hyperv"

	"github.com/code-ready/crc/pkg/crc/machine/config"

	winnet "github.com/code-ready/crc/pkg/os/windows/network"
)

func CreateHost(machineConfig config.MachineConfig) *hyperv.Driver {
	hypervDriver := hyperv.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, hypervDriver.VMDriver)

	hypervDriver.DisableDynamicMemory = true

	_, switchName := winnet.SelectSwitchByNameOrDefault(AlternativeNetwork)
	hypervDriver.VirtualSwitch = switchName

	return hypervDriver
}
