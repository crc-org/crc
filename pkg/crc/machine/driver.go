package machine

import (
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
