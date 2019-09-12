package machine

import (
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	HyperVDriver := MachineDriver{
		Name:     "Microsoft Hyper-V",
		Platform: crcos.WINDOWS,
		Driver:   "hyperv",
	}

	VirtualBoxWindowsDriver := MachineDriver{
		Name:     "VirtualBox",
		Platform: crcos.WINDOWS,
		Driver:   "virtualbox",
	}

	SupportedDrivers = []MachineDriver{
		HyperVDriver,
		VirtualBoxWindowsDriver,
	}

	DefaultDriver = HyperVDriver
}
