package constants

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	OcExecutableName   = "oc"
	TrayExecutableName = "CodeReady Containers.app"
)

var (
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
)

func TrayExecutablePath() string {
	if version.IsInstaller() {
		return filepath.Clean(filepath.Join(version.InstallPath(), "..", "MacOS", "CodeReady Containers"))
	}
	// Should not be reached, tray is only supported on installer builds
	return filepath.Clean(filepath.Join(BinDir(), "CodeReady Containers"))
}
