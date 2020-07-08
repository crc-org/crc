package cache

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
)

func NewMachineDriverHyperkitCache(version string, getVersion func() (string, error)) *Cache {
	return New(hyperkit.MachineDriverCommand, hyperkit.MachineDriverDownloadUrl, constants.CrcBinDir, version, getVersion)
}

func NewHyperkitCache(version string, getVersion func() (string, error)) *Cache {
	return New(hyperkit.HyperkitCommand, hyperkit.HyperkitDownloadUrl, constants.CrcBinDir, version, getVersion)
}
