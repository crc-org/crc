package constants

import (
	"github.com/code-ready/crc/pkg/crc/constants/windows"
)

const (
	OcBinaryName     = "oc.exe"
	PodmanBinaryName = "podman.exe"
)

func GetDefaultBundle() string {
	return windows.DefaultBundle
}

func GetOcUrl() string {
	return windows.OcUrl
}

func GetPodmanUrl() string {
	return windows.PodmanUrl
}
