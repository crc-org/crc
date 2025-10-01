//go:build windows

package fs9p

import (
	"fmt"
	"net"

	"github.com/linuxkit/virtsock/pkg/hvsock"
)

// GetHvsockListener returns a net.Listener listening on the specified hvsock.
// The used hvsock must already be defined before this function is called.
func GetHvsockListener(hvsockGUID string) (net.Listener, error) {
	service, err := hvsock.GUIDFromString(hvsockGUID)
	if err != nil {
		return nil, fmt.Errorf("parsing hvsock guid %s: %w", hvsockGUID, err)
	}

	listener, err := hvsock.Listen(hvsock.Addr{
		VMID:      hvsock.GUIDWildcard,
		ServiceID: service,
	})
	if err != nil {
		return nil, fmt.Errorf("retrieving listener for hvsock %s: %w", hvsockGUID, err)
	}

	return listener, nil
}
