package hyperv

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/machine/drivers/hyperv"

	"github.com/code-ready/crc/pkg/crc/machine/config"
)

func CreateHost(machineConfig config.MachineConfig) *hyperv.Driver {
	hypervDriver := hyperv.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	hypervDriver.CPU = machineConfig.CPUs
	hypervDriver.BundleName = machineConfig.BundleName
	hypervDriver.Memory = machineConfig.Memory

	// TODO: VirtualSwitch name is supposed to be localized or by ID
	hypervDriver.VirtualSwitch = "Default Switch"

	// DiskPath should come from the bundle's metadata (unflattened)
	// This force to add entry of DiskPath under crc machine config.json
	hypervDriver.DiskPath = filepath.Join(constants.MachineBaseDir, "machines", machineConfig.Name, "crc.vhdx")
	// The DiskPathURL will contain the cache path of where the image is located (not the actual image location)
	hypervDriver.DiskPathUrl = machineConfig.DiskPathURL
	hypervDriver.SSHKeyPath = machineConfig.SSHKeyPath

	return hypervDriver
}
