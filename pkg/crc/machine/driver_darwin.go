package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	"github.com/code-ready/crc/pkg/crc/machine/virtualbox"
)

func init() {
	HyperkitDriver := MachineDriver{
		Name:       "Hyperkit",
		Platform:   crcos.DARWIN,
		Driver:     "hyperkit",
		DriverPath: constants.CrcBinDir,
	}

	VirtualBoxMacOSDriver := MachineDriver{
		Name:     "VirtualBox",
		Platform: crcos.DARWIN,
		Driver:   "virtualbox",
	}

	SupportedDrivers = []MachineDriver{
		HyperkitDriver,
		VirtualBoxMacOSDriver,
	}

	DefaultDriver = HyperkitDriver
}

func getDriverOptions(machineConfig config.MachineConfig) interface{} {
	var driver interface{}

	// Supported drivers
	switch machineConfig.VMDriver {

	case "virtualbox":
		logging.Warn("Virtualbox support is deprecated and will be removed in the next release.")
		driver = virtualbox.CreateHost(machineConfig)
	case "hyperkit":
		driver = hyperkit.CreateHost(machineConfig)

	default:
		errors.ExitWithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
