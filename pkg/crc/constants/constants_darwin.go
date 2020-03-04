package constants

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants/darwin"
)

const (
	OcBinaryName     = "oc"
	PodmanBinaryName = "podman"
	TrayBinaryName   = "CodeReady Containers.app"
)

var (
	TrayAppBundlePath = filepath.Join(CrcBinDir, TrayBinaryName)
	TrayBinaryPath    = filepath.Join(TrayAppBundlePath, "Contents", "MacOS", "CodeReady Containers")
)

func GetDefaultBundle() string {
	return darwin.DefaultBundle
}

func GetOcUrl() string {
	return darwin.OcUrl
}

func GetPodmanUrl() string {
	return darwin.PodmanUrl
}
