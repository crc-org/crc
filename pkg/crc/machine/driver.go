package machine

import (
	crcos "github.com/code-ready/crc/pkg/os"
)

type Driver struct {
	Name       string
	Platform   crcos.OS
	Driver     string
	DriverPath string
}

var (
	SupportedDrivers []Driver
	DefaultDriver    Driver
)

func SupportedDriverValues() []string {
	supportedDrivers := []string{}
	for _, d := range SupportedDrivers {
		supportedDrivers = append(supportedDrivers, d.Driver)
	}
	return supportedDrivers
}
