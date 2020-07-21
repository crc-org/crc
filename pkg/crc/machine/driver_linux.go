package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
)

func init() {
	LibvirtDriver := Driver{
		Name:       "Libvirt",
		Platform:   crcos.LINUX,
		Driver:     "libvirt",
		DriverPath: constants.CrcBinDir,
	}

	SupportedDrivers = []Driver{
		LibvirtDriver,
	}

	DefaultDriver = LibvirtDriver
}

func getDriverOptions(machineConfig config.MachineConfig) interface{} {
	var driver interface{}

	// Supported drivers
	switch machineConfig.VMDriver {

	case "libvirt":
		driver = libvirt.CreateHost(machineConfig)

	default:
		errors.ExitWithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
