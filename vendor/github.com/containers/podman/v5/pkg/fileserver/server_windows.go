package fileserver

import (
	"fmt"

	"github.com/linuxkit/virtsock/pkg/hvsock"
	"github.com/sirupsen/logrus"
)

// Start serving the given shares on Windows HVSocks for use by a Hyper-V VM.
// Mounts is formatted as a map of directory to be shared to vsock GUID.
// The vsocks used must already be defined before StartShares is called; it's
// expected that the vsocks will be created and torn down by the program calling
// gvproxy.
// TODO: The map here probably doesn't make sense.
func StartHvsockShares(mounts map[string]string) (defErr error) {
	plan9Mounts := []Mount{}
	for path, guid := range mounts {
		service, err := hvsock.GUIDFromString(guid)
		if err != nil {
			return fmt.Errorf("parsing vsock guid %s: %w", guid, err)
		}

		listener, err := hvsock.Listen(hvsock.Addr{
			VMID:      hvsock.GUIDWildcard,
			ServiceID: service,
		})
		if err != nil {
			return fmt.Errorf("retrieving listener for vsock %s: %w", guid, err)
		}

		logrus.Debugf("Going to serve directory %s on vsock %s", path, guid)
		plan9Mounts = append(plan9Mounts, Mount{Listener: listener, Path: path})
	}
	return StartShares(plan9Mounts)
}
