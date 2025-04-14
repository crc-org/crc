package cache

import (
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/machine/vfkit"
)

func NewVfkitCache() *Cache {
	return newCache(vfkit.ExecutablePath(), vfkit.VfkitDownloadURL, vfkit.VfkitVersion, getVfkitVersion)
}

func getVfkitVersion(executablePath string) (string, error) {
	version, err := getVersionGeneric(executablePath, "-v")
	if err != nil {
		return version, err
	}
	return strings.TrimPrefix(version, "v"), nil
}
