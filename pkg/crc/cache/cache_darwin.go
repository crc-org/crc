package cache

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
)

func NewMachineDriverHyperkitCache() *Cache {
	return New(hyperkit.MachineDriverCommand, hyperkit.MachineDriverDownloadURL, constants.CrcBinDir, hyperkit.MachineDriverVersion, getHyperKitMachineDriverVersion)
}

func NewHyperkitCache() *Cache {
	return New(hyperkit.HyperkitCommand, hyperkit.HyperkitDownloadURL, constants.CrcBinDir, "", nil)
}

func getHyperKitMachineDriverVersion(binaryPath string) (string, error) {
	return getVersionGeneric(binaryPath, "version")
}
