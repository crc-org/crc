// +build !linux

package machine

import (
	"fmt"
	"runtime"
)

func copyDiskImage(dirName string, srcDir string) (string, string, error) {
	return "", "", fmt.Errorf("Not implemented for %s", runtime.GOOS)
}
