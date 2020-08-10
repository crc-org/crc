package hyperkit

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/machine/drivers/hyperkit"
)

func CreateHost(machineConfig config.MachineConfig) *hyperkit.Driver {
	hyperkitDriver := hyperkit.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	hyperkitDriver.BundleName = machineConfig.BundleName
	hyperkitDriver.Memory = machineConfig.Memory
	hyperkitDriver.CPU = machineConfig.CPUs
	hyperkitDriver.UUID = "c3d68012-0208-11ea-9fd7-f2189899ab08"
	hyperkitDriver.Cmdline = machineConfig.KernelCmdLine
	hyperkitDriver.VmlinuzPath = machineConfig.Kernel
	hyperkitDriver.InitrdPath = machineConfig.Initramfs
	hyperkitDriver.ImageSourcePath = machineConfig.ImageSourcePath
	hyperkitDriver.ImageFormat = machineConfig.ImageFormat
	hyperkitDriver.SSHKeyPath = machineConfig.SSHKeyPath
	hyperkitDriver.HyperKitPath = filepath.Join(constants.CrcBinDir, "hyperkit")

	return hyperkitDriver
}
