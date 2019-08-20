package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	LibvirtDriver := MachineDriver{
		Name:          "Libvirt",
		Platform:      crcos.LINUX,
		Driver:        "libvirt",
		UseDNSService: false,
		DriverPath:    constants.CrcBinDir,
	}

	VirtualBoxLinuxDriver := MachineDriver{
		Name:          "VirtualBox",
		Platform:      crcos.LINUX,
		Driver:        "virtualbox",
		UseDNSService: true,
	}

	SupportedDrivers = []MachineDriver{
		LibvirtDriver,
		VirtualBoxLinuxDriver,
	}

	DefaultDriver = LibvirtDriver
}
