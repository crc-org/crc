package constants

import "path/filepath"

const (
	OcExecutableName        = "oc"
	PodmanExecutableName    = "podman"
	TrayExecutableName      = "CodeReady Containers.app"
	GoodhostsExecutableName = "goodhosts"
)

var (
	TrayAppBundlePath  = filepath.Join(CrcBinDir, TrayExecutableName)
	TrayExecutablePath = filepath.Join(TrayAppBundlePath, "Contents", "MacOS", "CodeReady Containers")
)
