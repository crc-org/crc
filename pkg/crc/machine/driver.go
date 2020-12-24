package machine

import (
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/libmachine/host"
	libmachine "github.com/code-ready/machine/libmachine/drivers"
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

func setMemory(host *host.Host, memorySize int) error {
	memorySetter := func(driver *libmachine.VMDriver) bool {
		if driver.Memory == memorySize {
			return false
		}
		driver.Memory = memorySize
		return true
	}

	return updateDriverValue(host, memorySetter)
}

func setVcpus(host *host.Host, vcpus int) error {
	vcpuSetter := func(driver *libmachine.VMDriver) bool {
		if driver.CPU == vcpus {
			return false
		}
		driver.CPU = vcpus
		return true
	}

	return updateDriverValue(host, vcpuSetter)
}

func setDiskSize(host *host.Host, diskSizeGiB int) error {
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
