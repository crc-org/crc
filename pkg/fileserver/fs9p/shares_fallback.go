//go:build !windows

package fs9p

import (
	"fmt"
	"runtime"
)

// StartHvsockShares is only supported on Windows
func StartHvsockShares(mounts []HvsockMount9p) ([]*Server, error) {
	return nil, fmt.Errorf("StartHvsockShares() not implemented on %s", runtime.GOOS)
}
