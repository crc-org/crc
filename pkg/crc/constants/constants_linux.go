package constants

import (
	"github.com/code-ready/crc/pkg/crc/constants/linux"
)

const (
	OcBinaryName     = "oc"
	PodmanBinaryName = "podman"
)

func GetDefaultBundle() string {
	return linux.DefaultBundle
}

func GetOcUrl() string {
	return linux.OcUrl
}

func GetPodmanUrl() string {
	return linux.PodmanUrl
}
