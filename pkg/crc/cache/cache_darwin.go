package cache

import (
	"github.com/crc-org/crc/pkg/crc/machine/vfkit"
)

func NewVfkitCache() *Cache {
	return newCache(vfkit.ExecutablePath(), vfkit.VfkitDownloadURL, vfkit.VfkitVersion, getVfkitVersion)
}

func getVfkitVersion(executablePath string) (string, error) {
	return getVersionGeneric(executablePath, "-v")
}
