package constants

import "path/filepath"

const (
	OcBinaryName     = "oc"
	PodmanBinaryName = "podman"
	TrayBinaryName   = "CodeReady Containers.app"
)

var (
	TrayBinaryPath = filepath.Join(CrcBinDir, TrayBinaryName)
)
