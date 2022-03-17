package constants

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/version"
)

const (
	OcExecutableName           = "oc"
	PodmanRemoteExecutableName = "podman"
	TrayExecutableName         = "CodeReady Containers.app"
	DaemonAgentLabel           = "com.redhat.crc.daemon"
)

var (
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
)

func TrayExecutablePath() string {
	if version.IsInstaller() {
		return filepath.Clean(filepath.Join(version.InstallPath(), "..", "MacOS", "crc-tray"))
	}
	// Should not be reached, tray is only supported on installer builds
	return filepath.Clean(filepath.Join(BinDir(), "CodeReady Containers"))
}
