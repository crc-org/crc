package machine

import (
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperv"
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	HyperVDriver := Driver{
		Name:     "Microsoft Hyper-V",
		Platform: crcos.WINDOWS,
		Driver:   "hyperv",
	}

	SupportedDrivers = []Driver{
		HyperVDriver,
	}

	DefaultDriver = HyperVDriver
}

func getDriverOptions(machineConfig config.MachineConfig) interface{} {
	var driver interface{}

	// Supported drivers
	switch machineConfig.VMDriver {

	case "hyperv":
		driver = hyperv.CreateHost(machineConfig)

	default:
		exit.WithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
