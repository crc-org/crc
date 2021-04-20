package hyperkit

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/drivers/hyperkit"
)

func CreateHost(machineConfig config.MachineConfig) *hyperkit.Driver {
	hyperkitDriver := hyperkit.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, hyperkitDriver.VMDriver)

	hyperkitDriver.UUID = "c3d68012-0208-11ea-9fd7-f2189899ab08"
	hyperkitDriver.Cmdline = machineConfig.KernelCmdLine
	hyperkitDriver.VmlinuzPath = machineConfig.Kernel
	hyperkitDriver.InitrdPath = machineConfig.Initramfs
	hyperkitDriver.HyperKitPath = filepath.Join(constants.BinDir(), HyperKitCommand)

	hyperkitDriver.VMNet = machineConfig.NetworkMode == network.DefaultMode

	return hyperkitDriver
}
