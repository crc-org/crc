package config

import (
	"github.com/crc-org/machine/libmachine/drivers"
)

func InitVMDriverFromMachineConfig(machineConfig MachineConfig, driver *drivers.VMDriver) {
	driver.CPU = machineConfig.CPUs
	driver.Memory = uint(machineConfig.Memory)
	driver.DiskCapacity = uint64(machineConfig.DiskSize.ToBytes())
	driver.BundleName = machineConfig.BundleName
	driver.ImageSourcePath = machineConfig.ImageSourcePath
	driver.ImageFormat = machineConfig.ImageFormat
}
