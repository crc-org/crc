package constants

import "path/filepath"

const (
	OcBinaryName     = "oc"
	PodmanBinaryName = "podman"
	TrayBinaryName   = "CodeReady Containers.app"
)

var (
	TrayAppBundlePath = filepath.Join(CrcBinDir, TrayBinaryName)
	TrayBinaryPath    = filepath.Join(TrayAppBundlePath, "Contents", "MacOS", "CodeReady Containers")
)
