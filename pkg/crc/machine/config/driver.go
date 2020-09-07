package config

import (
	"github.com/code-ready/machine/libmachine/drivers"
)

func InitVMDriverFromMachineConfig(machineConfig MachineConfig, driver *drivers.VMDriver) {
	driver.CPU = machineConfig.CPUs
	driver.Memory = machineConfig.Memory
	driver.BundleName = machineConfig.BundleName
	driver.ImageSourcePath = machineConfig.ImageSourcePath
	driver.ImageFormat = machineConfig.ImageFormat
	driver.SSHKeyPath = machineConfig.SSHKeyPath
}
