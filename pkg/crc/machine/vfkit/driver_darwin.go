package vfkit

import (
	"fmt"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/drivers/vfkit"
	"github.com/code-ready/machine/libmachine/drivers"
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

	vfDriver.SharedDirs = configureShareDirs(machineConfig)

	return vfDriver
}

func configureShareDirs(machineConfig config.MachineConfig) []drivers.SharedDir {
	var sharedDirs []drivers.SharedDir
	for i, dir := range machineConfig.SharedDirs {
		sharedDir := drivers.SharedDir{
			Source: dir,
			Target: dir,
			Tag:    fmt.Sprintf("dir%d", i),
			Type:   "virtiofs",
		}
		sharedDirs = append(sharedDirs, sharedDir)
	}
	return sharedDirs
}
