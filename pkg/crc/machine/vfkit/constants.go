//go:build darwin || build
// +build darwin build

package vfkit

import (
	"fmt"
)

const (
	VfkitVersion = "0.0.2"
	VfkitCommand = "vfkit"
)

var (
	VfkitDownloadURL = fmt.Sprintf("https://github.com/code-ready/vfkit/releases/download/v%s/%s", VfkitVersion, VfkitCommand)
)
