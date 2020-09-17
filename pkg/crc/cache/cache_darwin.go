package cache

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
)

func NewMachineDriverHyperKitCache() *Cache {
	return New(hyperkit.MachineDriverCommand, hyperkit.MachineDriverDownloadURL, constants.CrcBinDir, hyperkit.MachineDriverVersion, getHyperKitMachineDriverVersion)
}

func NewHyperKitCache() *Cache {
	return New(hyperkit.HyperKitCommand, hyperkit.HyperKitDownloadURL, constants.CrcBinDir, "", nil)
}

func getHyperKitMachineDriverVersion(binaryPath string) (string, error) {
	return getVersionGeneric(binaryPath, "version")
}
