package machine

import (
	"github.com/code-ready/crc/pkg/crc/errors"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperv"
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
		errors.ExitWithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
