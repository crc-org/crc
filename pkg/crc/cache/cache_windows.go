package cache

import (
	"fmt"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/version"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
)

func NewWin32BackgroundLauncherCache() *Cache {
	url := constants.GetWin32BackgroundLauncherDownloadURL()
	version := version.GetWin32BackgroundLauncherVersion()
	cache := newCache(constants.Win32BackgroundLauncherPath(),
		url,
		version,
		func(executable string) (string, error) {
			stdOut, stdErr, err := powershell.Execute(fmt.Sprintf(`(Get-Item '%s').VersionInfo.FileVersion`, executable))
			if err != nil {
				return "", fmt.Errorf("unable to get version: %s: %w", stdErr, err)
			}
			return strings.TrimSpace(stdOut), nil
		},
	)
	// upstream binary name is win32-background-launcher.exe
	// but we want to name it as crc-background-launcher.exe
	cache.ignoreNameMismatch = true
	return cache
}
