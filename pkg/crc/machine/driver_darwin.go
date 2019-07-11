package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	HyperkitDriver := MachineDriver{
		Name:          "Hyperkit",
		Platform:      crcos.DARWIN,
		Driver:        "hyperkit",
		UseDNSService: true,
		DriverPath:    constants.CrcBinDir,
	}

	VirtualBoxMacOSDriver := MachineDriver{
		Name:          "VirtualBox",
		Platform:      crcos.DARWIN,
		Driver:        "virtualbox",
		UseDNSService: true,
	}

	SupportedDrivers = []MachineDriver{
		HyperkitDriver,
		VirtualBoxMacOSDriver,
	}

	DefaultDriver = HyperkitDriver
}
