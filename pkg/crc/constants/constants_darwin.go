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
	TrayExecutablePath   = filepath.Join(filepath.Dir(filepath.Dir(version.GetMacosInstallPath())), "MacOS", "CodeReady Containers")
	TapSocketPath        = filepath.Join(CrcBaseDir, "tap.sock")
	DaemonHTTPSocketPath = filepath.Join(CrcBaseDir, "crc-http.sock")
)
