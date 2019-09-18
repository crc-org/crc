package machine

import (
	"github.com/code-ready/crc/pkg/crc/errors"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperv"
	"github.com/code-ready/crc/pkg/crc/machine/virtualbox"
)

func init() {
	HyperVDriver := MachineDriver{
		Name:     "Microsoft Hyper-V",
		Platform: crcos.WINDOWS,
		Driver:   "hyperv",
	}

	VirtualBoxWindowsDriver := MachineDriver{
		Name:     "VirtualBox",
		Platform: crcos.WINDOWS,
		Driver:   "virtualbox",
	}

	SupportedDrivers = []MachineDriver{
		HyperVDriver,
		VirtualBoxWindowsDriver,
	}

	DefaultDriver = HyperVDriver
}

func getDriverOptions(machineConfig config.MachineConfig) interface{} {
	var driver interface{}

	// Supported drivers
	switch machineConfig.VMDriver {

	case "virtualbox":
		driver = virtualbox.CreateHost(machineConfig)
	case "hyperv":
		driver = hyperv.CreateHost(machineConfig)

	default:
		errors.ExitWithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
