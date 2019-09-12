package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
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
