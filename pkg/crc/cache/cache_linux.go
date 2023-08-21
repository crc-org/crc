package cache

import (
	"github.com/crc-org/crc/v2/pkg/crc/machine/libvirt"
)

func NewMachineDriverLibvirtCache() *Cache {
	return newCache(libvirt.MachineDriverPath(), libvirt.MachineDriverDownloadURL, libvirt.MachineDriverVersion, getCurrentLibvirtDriverVersion)
}

func getCurrentLibvirtDriverVersion(executablePath string) (string, error) {
	return getVersionGeneric(executablePath, "version")
}
