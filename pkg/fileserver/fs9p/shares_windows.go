package fs9p

import (
	"fmt"

	"github.com/linuxkit/virtsock/pkg/hvsock"
	"github.com/sirupsen/logrus"
)

// StartHvsockShares starts serving the given shares on hvsocks instead of TCP sockets.
// The hvsocks used must already be defined before StartHvsockShares is called.
func StartHvsockShares(mounts []HvsockMount9p) ([]*Server, error) {
	mounts9p := []Mount9p{}
	for _, mount := range mounts {
		service, err := hvsock.GUIDFromString(mount.HvsockGUID)
		if err != nil {
			return nil, fmt.Errorf("parsing hvsock guid %s: %w", mount.HvsockGUID, err)
		}

		listener, err := hvsock.Listen(hvsock.Addr{
			VMID:      hvsock.GUIDWildcard,
			ServiceID: service,
		})
		if err != nil {
			return nil, fmt.Errorf("retrieving listener for hvsock %s: %w", mount.HvsockGUID, err)
		}

		logrus.Debugf("Going to serve directory %s on hvsock %s", mount.Path, mount.HvsockGUID)
		mounts9p = append(mounts9p, Mount9p{Path: mount.Path, Listener: listener})
	}

	return StartShares(mounts9p)
}
