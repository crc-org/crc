package machine

import (
	"github.com/containers/common/pkg/strongunits"
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

func setMemory(host *host.Host, memorySize strongunits.MiB) error {
	memorySetter := func(driver *libmachine.VMDriver) bool {
		if driver.Memory == uint(memorySize) {
			return false
		}
		driver.Memory = uint(memorySize)
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

func setDiskSize(host *host.Host, diskSize strongunits.GiB) error {
	diskSizeSetter := func(driver *libmachine.VMDriver) bool {
		capacity := diskSize.ToBytes()
		if driver.DiskCapacity == uint64(capacity) {
			return false
		}
		driver.DiskCapacity = uint64(capacity)
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
