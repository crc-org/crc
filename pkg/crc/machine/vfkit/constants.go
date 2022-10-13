//go:build darwin || build
// +build darwin build

package vfkit

import (
	"fmt"
)

const (
	VfkitVersion = "0.0.4"
	VfkitCommand = "vfkit"
)

var (
	VfkitDownloadURL = fmt.Sprintf("https://github.com/crc-org/vfkit/releases/download/v%s/%s", VfkitVersion, VfkitCommand)
)
