package plan9

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

// from /etc/services:
// 9pfs            564/tcp                 # plan 9 file service
const Port = 564
const PortStr = "564"

type Mount struct {
	Listener net.Listener
	Path     string
}

func StartShares(plan9Mounts []Mount) (defErr error) {
	for _, m := range plan9Mounts {
		server, err := New9pServer(m.Listener, m.Path)
		if err != nil {
			return fmt.Errorf("serving directory %s on vsock %s: %w", m.Path, m.Listener.Addr().String(), err)
		}
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
				logrus.Errorf("Error from 9p server for %s: %v", serverDir, err)
			} else {
				// We do not expect server exits - this should
				// run until the program exits.
				logrus.Warnf("9p server for %s exited without error", serverDir)
			}
		}()
	}

	return nil
}
