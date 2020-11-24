package constants

import "path/filepath"

const (
	OcExecutableName          = "oc"
	PodmanExecutableName      = "podman"
	TrayExecutableName        = "CodeReady Containers.app"
	AdminHelperExecutableName = "admin-helper"
)

var (
	TrayAppBundlePath  = filepath.Join(CrcBinDir, TrayExecutableName)
	TrayExecutablePath = filepath.Join(TrayAppBundlePath, "Contents", "MacOS", "CodeReady Containers")
)
