package machine

import (
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	HyperVDriver := MachineDriver{
		Name:          "Microsoft Hyper-V",
		Platform:      crcos.WINDOWS,
		Driver:        "hyperv",
		UseDNSService: true,
	}

	VirtualBoxWindowsDriver := MachineDriver{
		Name:          "VirtualBox",
		Platform:      crcos.WINDOWS,
		Driver:        "virtualbox",
		UseDNSService: true,
	}

	SupportedDrivers = []MachineDriver{
		HyperVDriver,
		VirtualBoxWindowsDriver,
	}

	DefaultDriver = HyperVDriver
}
