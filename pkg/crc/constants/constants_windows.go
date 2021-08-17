package constants

import (
	"path/filepath"
)

const (
	OcExecutableName    = "oc.exe"
	TrayExecutableName  = "crc-tray.exe"
	TrayShortcutName    = "crc-tray.lnk"
	TapSocketPath       = ""
	DaemonHTTPNamedPipe = `\\.\pipe\crc-http`
)

var (
	TrayExecutablePath = filepath.Join(BinDir(), TrayExecutableName)
)
