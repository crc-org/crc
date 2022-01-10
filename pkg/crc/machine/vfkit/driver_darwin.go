package vfkit

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/drivers/vfkit"
)

func CreateHost(machineConfig config.MachineConfig) *vfkit.Driver {
	vfDriver := vfkit.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, vfDriver.VMDriver)

	vfDriver.Cmdline = machineConfig.KernelCmdLine
	vfDriver.VmlinuzPath = machineConfig.Kernel
	vfDriver.InitrdPath = machineConfig.Initramfs
	vfDriver.VfkitPath = filepath.Join(constants.BinDir(), VfkitCommand)

	vfDriver.VirtioNet = machineConfig.NetworkMode == network.SystemNetworkingMode
	vfDriver.VsockPath = constants.TapSocketPath

	return vfDriver
}
