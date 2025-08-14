package fs9p

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/DeedleFake/p9"
	"github.com/DeedleFake/p9/proto"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/sirupsen/logrus"
)

type Server struct {
	// Listener this server is bound to
	Listener net.Listener
	// Directory this server exposes
	ExposedDir string
	// Errors from the server being started will come out here
	ErrChan chan error
}

// New9pServer exposes a single directory (and all children) via the given net.Listener.
// Directory given must be an absolute path and must exist.
func New9pServer(listener net.Listener, exposeDir string) (*Server, error) {
	// Verify that exposeDir makes sense.
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

	errChan := make(chan error)

	fs := p9.FileSystem(p9.Dir(exposeDir))

	go func() {
		errChan <- proto.Serve(listener, p9.Proto(), p9.FSConnHandler(fs, constants.Plan9Msize))
		close(errChan)
	}()

	toReturn := new(Server)
	toReturn.Listener = listener
	toReturn.ExposedDir = exposeDir
	toReturn.ErrChan = errChan

	// Just before returning, check to see if we got an error off server startup.
	select {
	case err := <-errChan:
		return nil, fmt.Errorf("starting 9p server: %w", err)
	default:
		logrus.Infof("Successfully started 9p server on %s for directory %s", listener.Addr().String(), exposeDir)
	}

	return toReturn, nil
}

// Stop a running server.
// Please note that this does *BAD THINGS* to clients if they are still running
// when the server stops. Processes get stuck in I/O deep sleep and zombify, and
// nothing I do other than restarting the VM can remove the zombies.
func (s *Server) Stop() error {
	if err := s.Listener.Close(); err != nil {
		return err
	}
	logrus.Infof("Successfully stopped 9p server for directory %s", s.ExposedDir)
	return nil
}

// WaitForError from a running server.
func (s *Server) WaitForError() error {
	err := <-s.ErrChan
	return err
}
