//go:build darwin || build

package vfkit

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
)

const (
	VfkitVersion   = "0.6.1"
	VfkitCommand   = "vfkit"
	KrunkitVersion = "1.1.1"
	KrunkitCommand = "krunkit"
)

var (
	VfkitDownloadURL     = fmt.Sprintf("https://github.com/crc-org/vfkit/releases/download/v%s/%s", VfkitVersion, VfkitCommand)
	VfkitEntitlementsURL = fmt.Sprintf("https://raw.githubusercontent.com/crc-org/vfkit/v%s/vf.entitlements", VfkitVersion)
	KrunkitDownloadURL   = fmt.Sprintf("https://github.com/containers/krunkit/releases/download/v%s/krunkit-podman-unsigned-%s.tgz", KrunkitVersion, KrunkitVersion)
)

func ExecutablePath() string {
	return constants.ResolveHelperPath(constants.Provider())
}

func DownloadURL() string {
	switch constants.Provider() {
	case VfkitCommand:
		return VfkitDownloadURL
	case KrunkitCommand:
		return KrunkitDownloadURL
	default:
		return ""
	}
}

func Version() string {
	switch constants.Provider() {
	case VfkitCommand:
		return VfkitVersion
	case KrunkitCommand:
		return KrunkitVersion
	default:
		return ""
	}
}
