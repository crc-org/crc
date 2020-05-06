package constants

import "path/filepath"

const (
	OcBinaryName        = "oc.exe"
	PodmanBinaryName    = "podman.exe"
	GoodhostsBinaryName = "goodhosts.exe"
	TrayBinaryName      = "tray-windows.exe"
	DaemonServiceName   = "CodeReady Containers"
	TrayShortcutName    = "tray-windows.lnk"
)

var TrayBinaryPath = filepath.Join(CrcBinDir, "tray-windows", TrayBinaryName)
