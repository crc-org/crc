package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	HyperkitDriver := Driver{
		Name:       "Hyperkit",
		Platform:   crcos.DARWIN,
		Driver:     "hyperkit",
		DriverPath: constants.CrcBinDir,
	}

	SupportedDrivers = []Driver{
		HyperkitDriver,
	}

	DefaultDriver = HyperkitDriver
}

func getDriverOptions(machineConfig config.MachineConfig) interface{} {
	var driver interface{}

	// Supported drivers
	switch machineConfig.VMDriver {

	case "hyperkit":
		driver = hyperkit.CreateHost(machineConfig)

	default:
		exit.WithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
