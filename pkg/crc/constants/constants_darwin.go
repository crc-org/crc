package constants

import "path/filepath"

const (
	OcBinaryName   = "oc"
	TrayBinaryName = "CodeReady Containers.app"
)

var (
	TrayBinaryPath = filepath.Join(CrcBinDir, TrayBinaryName)
)
