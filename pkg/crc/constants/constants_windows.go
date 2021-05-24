package constants

import (
	"os"
	"path/filepath"
)

const (
	OcExecutableName            = "oc.exe"
	AdminHelperExecutableName   = "admin-helper-windows.exe"
	TrayExecutableName          = "crc-tray.exe"
	TrayShortcutName            = "crc-tray.lnk"
	DaemonBatchFileName         = "crc-daemon-autostart.bat"
	DaemonPSScriptName          = "launch-crc-daemon.ps1"
	DaemonBatchFileShortcutName = "crc-daemmon-autostart.bat.lnk"
	TapSocketPath               = ""
	DaemonHTTPNamedPipe         = `\\.\pipe\crc-http`
)

var (
	StartupFolder       = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	TrayExecutablePath  = filepath.Join(BinDir(), TrayExecutableName)
	DaemonBatchFilePath = filepath.Join(CrcBinDir, DaemonBatchFileName)
	DaemonPSScriptPath  = filepath.Join(CrcBinDir, DaemonPSScriptName)
)
