package constants

import (
	"os"
	"path/filepath"
)

const (
	OcExecutableName            = "oc.exe"
	PodmanExecutableName        = "podman.exe"
	AdminHelperExecutableName   = "admin-helper-windows.exe"
	TrayExecutableName          = "tray-windows.exe"
	TrayShortcutName            = "tray-windows.lnk"
	DaemonBatchFileName         = "crc-daemon-autostart.bat"
	DaemonPSScriptName          = "launch-crc-daemon.ps1"
	DaemonBatchFileShortcutName = "crc-daemmon-autostart.bat.lnk"
	TapSocketPath               = ""
	DaemonHTTPNamedPipe         = `\\.\pipe\crc-http`
)

var (
	StartupFolder       = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	TrayExecutableDir   = filepath.Join(CrcBinDir, "tray-windows")
	TrayExecutablePath  = filepath.Join(TrayExecutableDir, TrayExecutableName)
	DaemonBatchFilePath = filepath.Join(CrcBinDir, DaemonBatchFileName)
	DaemonPSScriptPath  = filepath.Join(CrcBinDir, DaemonPSScriptName)
)
