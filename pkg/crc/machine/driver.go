package machine

import (
	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	"github.com/crc-org/crc/v2/pkg/libmachine/host"
	libmachine "github.com/crc-org/machine/libmachine/drivers"
)

type valueSetter func(driver *libmachine.VMDriver) bool

func updateDriverValue(host *host.Host, setDriverValue valueSetter) error {
	driver, err := loadDriverConfig(host)
	if err != nil {
		return err
	}
	valueChanged := setDriverValue(driver.VMDriver)
	if !valueChanged {
		return nil
	}

	return updateDriverConfig(host, driver)
}

func setMemory(host *host.Host, memorySize uint) error {
	memorySetter := func(driver *libmachine.VMDriver) bool {
		if driver.Memory == memorySize {
			return false
		}
		driver.Memory = memorySize
		return true
	}

	return updateDriverValue(host, memorySetter)
}

func setVcpus(host *host.Host, vcpus uint) error {
	vcpuSetter := func(driver *libmachine.VMDriver) bool {
		if driver.CPU == vcpus {
			return false
		}
		driver.CPU = vcpus
		return true
	}

	return updateDriverValue(host, vcpuSetter)
}

func setDiskSize(host *host.Host, diskSizeGiB uint) error {
	diskSizeSetter := func(driver *libmachine.VMDriver) bool {
		capacity := config.ConvertGiBToBytes(diskSizeGiB)
		if driver.DiskCapacity == capacity {
			return false
		}
		driver.DiskCapacity = capacity
		return true
	}

	return updateDriverValue(host, diskSizeSetter)
}

func setSharedDirPassword(host *host.Host, password string) error {
	driver, err := loadDriverConfig(host)
	if err != nil {
		return err
	}

	if len(driver.SharedDirs) == 0 {
		return nil
	}

	for i := range driver.SharedDirs {
		driver.SharedDirs[i].Password = password
	}
	return updateDriverStruct(host, driver)
}
