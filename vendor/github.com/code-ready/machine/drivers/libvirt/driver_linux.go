package libvirt

import (
	"github.com/code-ready/machine/libmachine/drivers"
)

type Driver struct {
	*drivers.VMDriver

	// Driver specific configuration
	Network     string
	CacheMode   string
	IOMode      string
	VSock       bool
	StoragePool string
}

const (
	defaultMemory    = 8192
	defaultCPU       = 4
	defaultCacheMode = "default"
	defaultIOMode    = "threads"
)

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		VMDriver: &drivers.VMDriver{
			BaseDriver: &drivers.BaseDriver{
				MachineName: hostName,
				StorePath:   storePath,
			},
			Memory: defaultMemory,
			CPU:    defaultCPU,
		},
		CacheMode: defaultCacheMode,
		IOMode:    defaultIOMode,
	}
}
