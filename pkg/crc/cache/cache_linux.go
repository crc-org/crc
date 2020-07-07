package cache

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
)

func NewMachineDriverLibvirtCache(version string, getVersion func() (string, error)) *Cache {
	return New(libvirt.MachineDriverCommand, libvirt.MachineDriverDownloadUrl, constants.CrcBinDir, version, getVersion)
}
