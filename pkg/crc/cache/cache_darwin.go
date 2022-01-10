package cache

import (
	"github.com/code-ready/crc/pkg/crc/machine/vfkit"
)

func NewVfkitCache() *Cache {
	return New(vfkit.VfkitCommand, vfkit.VfkitDownloadURL, vfkit.VfkitVersion, getVfkitVersion)
}

func getVfkitVersion(executablePath string) (string, error) {
	return getVersionGeneric(executablePath, "-v")
}
