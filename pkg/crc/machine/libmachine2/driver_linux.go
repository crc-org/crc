package libmachine2

import (
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/machine/libmachine/drivers"
)

func driver(machineConfig config.MachineConfig, baseDriver *drivers.BaseDriver) drivers.Driver {
	return libvirt.CreateHost(machineConfig, baseDriver)
}
