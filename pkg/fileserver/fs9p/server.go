package fs9p

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeedleFake/p9"
	"github.com/DeedleFake/p9/proto"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/sirupsen/logrus"
)

type Server struct {
	// Listener this server is bound to
	Listener net.Listener

	// Plan9 Filesystem type that holds the exposed directory
	Filesystem p9.FileSystem

	// Directory this server exposes
	ExposedDir string

	// Errors from the server being started will come out here
	ErrChan chan error
}

// New9pServer exposes a single directory (and all children) via the given net.Listener
// and returns the server struct.
// Directory given must be an absolute path and must exist.
func New9pServer(listener net.Listener, exposeDir string) (*Server, error) {
	// verify that exposeDir makes sense
	if !filepath.IsAbs(exposeDir) {
		return nil, fmt.Errorf("path to expose to machine must be absolute: %s", exposeDir)
	}
	stat, err := os.Stat(exposeDir)
	if err != nil {
		return nil, fmt.Errorf("cannot stat path to expose to machine: %w", err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("path to expose to machine must be a directory: %s", exposeDir)
	}

	fs := p9.FileSystem(p9.Dir(exposeDir))
	// set size to 1 making channel buffered to prevent proto.Serve blocking
	errChan := make(chan error, 1)

	toReturn := new(Server)
	toReturn.Listener = listener
	toReturn.Filesystem = fs
	toReturn.ExposedDir = exposeDir
	toReturn.ErrChan = errChan

	return toReturn, nil
}

// Start a server created by New9pServer.
func (s *Server) Start() error {
	go func() {
		s.ErrChan <- proto.Serve(s.Listener, p9.Proto(), p9.FSConnHandler(s.Filesystem, constants.Plan9Msize))
		close(s.ErrChan)
	}()

	// Just before returning, check to see if we got an error off server startup.
	select {
	case err := <-s.ErrChan:
		return fmt.Errorf("starting 9p server: %w", err)
	default:
		logrus.Infof("started 9p server on %s for directory %s", s.Listener.Addr().String(), s.ExposedDir)
		return nil
	}
}

// Stop a running server.
// Please note that this does *BAD THINGS* to clients if they are still running
// when the server stops. Processes get stuck in I/O deep sleep and zombify, and
// nothing I do other than restarting the VM can remove the zombies.
func (s *Server) Stop() error {
	if err := s.Listener.Close(); err != nil {
		return err
	}
	logrus.Infof("stopped 9p server for directory %s", s.ExposedDir)
	return nil
}

// WaitForError from a running server.
func (s *Server) WaitForError() error {
	err := <-s.ErrChan
	// captures "accept tcp: endpoint is in invalid state" errors on exit
	var opErr *net.OpError
	if errors.As(err, &opErr) && strings.Contains(opErr.Error(), "endpoint is in invalid state") {
		return nil
	}
	return err
}
