package machine

import (
	"github.com/code-ready/crc/pkg/crc/errors"

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
