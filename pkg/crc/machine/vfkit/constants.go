//go:build darwin || build

package vfkit

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
)

const (
	VfkitVersion = "1.1.1"
	vfkitCommand = "krunkit"
)

var (
	VfkitDownloadURL     = fmt.Sprintf("https://github.com/crc-org/vfkit/releases/download/v%s/%s", VfkitVersion, vfkitCommand)
	VfkitEntitlementsURL = fmt.Sprintf("https://raw.githubusercontent.com/crc-org/vfkit/v%s/vf.entitlements", VfkitVersion)
)

func ExecutablePath() string {
	return constants.ResolveHelperPath(vfkitCommand)
}
