package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
)

func init() {
	HyperkitDriver := MachineDriver{
		Name:       "Hyperkit",
		Platform:   crcos.DARWIN,
		Driver:     "hyperkit",
		DriverPath: constants.CrcBinDir,
	}

	SupportedDrivers = []MachineDriver{
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
		errors.ExitWithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
