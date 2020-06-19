package libmachine2

import (
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	"github.com/code-ready/machine/libmachine/drivers"
)

func driver(machineConfig config.MachineConfig) drivers.Driver {
	return hyperkit.CreateHost(machineConfig)
}
