package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

func init() {
	LibvirtDriver := MachineDriver{
		Name:       "Libvirt",
		Platform:   crcos.LINUX,
		Driver:     "libvirt",
		DriverPath: constants.CrcBinDir,
	}

	SupportedDrivers = []MachineDriver{
		LibvirtDriver,
	}

	DefaultDriver = LibvirtDriver
}
