package fs9p

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

// Mount9p represents an exposed directory path with the listener
// the 9p server is bound to
type Mount9p struct {
	Path     string
	Listener net.Listener
}

// VsockMount9p represents an exposed directory path with the Vsock GUID
// the 9p server is bound to
type VsockMount9p struct {
	Path      string
	VsockGUID string
}

// StartShares starts a new 9p server for each of the supplied mounts.
func StartShares(mounts []Mount9p) (servers []*Server, defErr error) {
	servers9p := []*Server{}
	for _, m := range mounts {
		server, err := New9pServer(m.Listener, m.Path)
		if err != nil {
			return nil, fmt.Errorf("serving directory %s on %s: %w", m.Path, m.Listener.Addr().String(), err)
		}

		servers9p = append(servers9p, server)

		defer func() {
			if defErr != nil {
				if err := server.Stop(); err != nil {
					logrus.Errorf("Error stopping 9p server: %v", err)
				}
			}
		}()

		serverDir := m.Path

		go func() {
			if err := server.WaitForError(); err != nil {
				logrus.Errorf("Error from 9p server on %s for %s: %v", m.Listener.Addr().String(), serverDir, err)
			} else {
				// We do not expect server exits - this should run until the program exits.
				logrus.Warnf("9p server on %s for %s exited without error", m.Listener.Addr().String(), serverDir)
			}
		}()
	}

	return servers9p, nil
}

// StartVsockShares starts serving the given shares on vsocks instead of TCP sockets.
// The vsocks used must already be defined before StartVsockShares is called.
func StartVsockShares(mounts []VsockMount9p) ([]*Server, error) {
	// temporarily removed until vsock support is added into 9pfs client
	return nil, nil
}
