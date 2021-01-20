package constants

import (
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	OcExecutableName          = "oc"
	PodmanExecutableName      = "podman"
	TrayExecutableName        = "CodeReady Containers.app"
	AdminHelperExecutableName = "admin-helper-darwin"
)

var (
	TrayAppBundlePath  = trayAppBundlePath()
	TrayExecutablePath = filepath.Join(TrayAppBundlePath, "Contents", "MacOS", "CodeReady Containers")
)

func trayAppBundlePath() string {
	if version.IsMacosInstallPathSet() {
		path := filepath.Join(version.GetMacosInstallPath(), version.GetCRCVersion(), TrayExecutableName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return filepath.Join(CrcBinDir, TrayExecutableName)
}
