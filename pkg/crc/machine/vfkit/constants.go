//go:build darwin || build
// +build darwin build

package vfkit

import (
	"fmt"
	"runtime"
)

const (
	VfkitVersion = "0.0.1"
)

var (
	VfkitCommand     = fmt.Sprintf("vfkit-%s", runtime.GOARCH)
	VfkitDownloadURL = fmt.Sprintf("https://github.com/code-ready/vfkit/releases/download/v%s/%s", VfkitVersion, VfkitCommand)
)
