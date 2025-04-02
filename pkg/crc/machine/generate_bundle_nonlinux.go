//go:build !linux
// +build !linux

package machine

import (
	"fmt"
	"runtime"
)

func copyDiskImage(_ string) (string, string, error) {
	return "", "", fmt.Errorf("not implemented for %s", runtime.GOOS)
}
