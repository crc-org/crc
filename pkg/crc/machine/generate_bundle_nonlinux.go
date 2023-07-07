//go:build !linux
// +build !linux

package machine

import (
	"fmt"
	"runtime"

	"github.com/crc-org/crc/v2/pkg/crc/preset"
)

func copyDiskImage(_ string, _ preset.Preset) (string, string, error) {
	return "", "", fmt.Errorf("Not implemented for %s", runtime.GOOS)
}
