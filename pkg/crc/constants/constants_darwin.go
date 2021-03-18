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
	TrayAppBundlePath    = trayAppBundlePath()
	TrayExecutablePath   = filepath.Join(TrayAppBundlePath, "Contents", "MacOS", "CodeReady Containers")
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
)

func trayAppBundlePath() string {
	if version.IsMacosInstallPathSet() {
		path := filepath.Join(filepath.Dir(filepath.Dir(version.GetMacosInstallPath())), TrayExecutableName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return filepath.Join(CrcBinDir, TrayExecutableName)
}
