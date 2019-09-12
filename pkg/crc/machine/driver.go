package machine

import (
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	"github.com/code-ready/crc/pkg/crc/machine/hyperv"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/crc/machine/virtualbox"
	crcos "github.com/code-ready/crc/pkg/os"
)

type MachineDriver struct {
	Name       string
	Platform   crcos.OS
	Driver     string
	DriverPath string
}

var (
	SupportedDrivers []MachineDriver
	DefaultDriver    MachineDriver
)

func SupportedDriverValues() []string {
	supportedDrivers := []string{}
	for _, d := range SupportedDrivers {
		supportedDrivers = append(supportedDrivers, d.Driver)
	}
	return supportedDrivers
}

func getDriverInfo(driver string) (*MachineDriver, error) {
	for _, d := range SupportedDrivers {
		if driver == d.Driver {
			return &d, nil
		}
	}
	return nil, errors.Newf("No info about unknown driver: %s", driver)
}

func getDriverOptions(machineConfig config.MachineConfig) interface{} {
	var driver interface{}

	// Supported drivers
	switch machineConfig.VMDriver {

	case "libvirt":
		driver = libvirt.CreateHost(machineConfig)
	case "virtualbox":
		driver = virtualbox.CreateHost(machineConfig)
	case "hyperkit":
		driver = hyperkit.CreateHost(machineConfig)
	case "hyperv":
		driver = hyperv.CreateHost(machineConfig)

	default:
		errors.ExitWithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}
