package machine

import (
	libmachine "github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/host"
)

type valueSetter func(driver *libmachine.VMDriver)

func updateDriverValue(host *host.Host, setDriverValue valueSetter) error {
	driver, err := loadDriverConfig(host)
	if err != nil {
		return err
	}
	setDriverValue(driver.VMDriver)

	return updateDriverConfig(host, driver)
}

func setMemory(host *host.Host, memorySize int) error {
	memorySetter := func(driver *libmachine.VMDriver) {
		driver.Memory = memorySize
	}

	return updateDriverValue(host, memorySetter)
}

func setVcpus(host *host.Host, vcpus int) error {
	vcpuSetter := func(driver *libmachine.VMDriver) {
		driver.CPU = vcpus
	}

	return updateDriverValue(host, vcpuSetter)
}
