package fileserver

import (
	"fmt"
	"net"

	"github.com/containers/podman/v5/pkg/fileserver/plan9"
	"github.com/sirupsen/logrus"
)

type Mount struct {
	Listener net.Listener
	Path     string
}

func StartShares(plan9Mounts []Mount) (defErr error) {
	for _, m := range plan9Mounts {
		server, err := plan9.New9pServer(m.Listener, m.Path)
		if err != nil {
			return fmt.Errorf("serving directory %s on %s: %w", m.Path, m.Listener.Addr().String(), err)
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
