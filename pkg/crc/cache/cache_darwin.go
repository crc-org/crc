package cache

import (
	"path/filepath"

	"github.com/crc-org/crc/pkg/crc/machine/vfkit"
)

func NewVfkitCache() *Cache {
	return New(filepath.Base(vfkit.ExecutablePath()), vfkit.VfkitDownloadURL, vfkit.VfkitVersion, getVfkitVersion)
}

func getVfkitVersion(executablePath string) (string, error) {
	return getVersionGeneric(executablePath, "-v")
}
