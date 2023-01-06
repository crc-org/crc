//go:build darwin || build
// +build darwin build

package vfkit

import (
	"fmt"
	"path/filepath"

	"github.com/crc-org/crc/pkg/crc/constants"
)

const (
	VfkitVersion = "0.0.4"
	vfkitCommand = "vfkit"
)

var (
	VfkitDownloadURL     = fmt.Sprintf("https://github.com/crc-org/vfkit/releases/download/v%s/%s", VfkitVersion, vfkitCommand)
	VfkitEntitlementsURL = fmt.Sprintf("https://raw.githubusercontent.com/crc-org/vfkit/v%s/vf.entitlements", VfkitVersion)
)

func ExecutablePath() string {
	return filepath.Join(constants.BinDir(), vfkitCommand)
}
