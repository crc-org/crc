package virtualbox

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/machine/drivers/virtualbox"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/machine/config"
)

func CreateHost(machineConfig config.MachineConfig) *virtualbox.Driver {
	virtualboxDriver := virtualbox.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	virtualboxDriver.CPU = machineConfig.CPUs
	virtualboxDriver.BundlePath = machineConfig.BundlePath
	virtualboxDriver.Memory = machineConfig.Memory

	// Network
	virtualboxDriver.HostOnlyCIDR = "192.168.130.1/24"

	// DiskPath should come from the bundle's metadata (unflattened)
	// This force to add entry of DiskPath under crc machine config.json
	virtualboxDriver.DiskPath = filepath.Join(constants.MachineBaseDir, "machines", machineConfig.Name, constants.DefaultDiskImage)
	virtualboxDriver.DiskPathUrl = machineConfig.DiskPathURL
	virtualboxDriver.SSHKeyPath = machineConfig.SSHKeyPath

	return virtualboxDriver
}
