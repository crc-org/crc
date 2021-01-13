package constants

import "path/filepath"

const (
	OcExecutableName          = "oc.exe"
	PodmanExecutableName      = "podman.exe"
	AdminHelperExecutableName = "admin-helper-windows.exe"
	TrayExecutableName        = "tray-windows.exe"
	DaemonServiceName         = "CodeReady Containers"
	TrayShortcutName          = "tray-windows.lnk"
)

var (
	TrayExecutableDir  = filepath.Join(CrcBinDir, "tray-windows")
	TrayExecutablePath = filepath.Join(TrayExecutableDir, TrayExecutableName)
)
