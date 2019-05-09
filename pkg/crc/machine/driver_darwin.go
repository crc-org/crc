package machine

import (
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	HyperkitDriver := MachineDriver{
		Name:          "Hyperkit",
		Platform:      crcos.DARWIN,
		Driver:        "hyperkit",
		UseDNSService: true,
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
