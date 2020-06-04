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

var (
	TrayBinaryDir  = filepath.Join(CrcBinDir, "tray-windows")
	TrayBinaryPath = filepath.Join(TrayBinaryDir, TrayBinaryName)
)
