// +build !linux

package machine

import (
	"fmt"
	"runtime"
)

func copyDiskImage(dirName string) (string, string, error) {
	return "", "", fmt.Errorf("Not implemented for %s", runtime.GOOS)
}
