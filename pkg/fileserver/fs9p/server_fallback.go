//go:build !windows

package fs9p

import (
	"fmt"
	"net"
	"runtime"
)

// GetHvsockListener returns a net.Listener listening on the specified hvsock.
// The used hvsock must already be defined before this function is called.
func GetHvsockListener(hvsockGUID string) (net.Listener, error) {
	return nil, fmt.Errorf("GetHvsockListener() not implemented on %s", runtime.GOOS)
}
