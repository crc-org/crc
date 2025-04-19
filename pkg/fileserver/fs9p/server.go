package fs9p

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/fs/real"
	"github.com/sirupsen/logrus"
)

type Server struct {
	server *go9p.Srv
	// TODO: Once server has a proper Close() we don't need this.
	// This is basically just a short-circuit to actually close the server
	// without that ability.
	listener net.Listener
	// Errors from the server being started will come out here.
	errChan chan error
}

// Modification of go9p's Serve that accepts a net.Listener as a parameter
// instead of an address.
func serveListener(listener net.Listener, srv go9p.Srv) error {
	for {
		a, err := listener.Accept()
		if err != nil {
			return err
		}
		go func(nc net.Conn, srv go9p.Srv) {
			defer nc.Close()
			read := bufio.NewReader(nc)
			err := go9p.ServeReadWriter(read, nc, srv)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		}(a, srv)
	}
}

// Expose a single directory (and all children) via the given net.Listener.
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

	filesys, _ := fs.NewFS("user", "user", 0777)
	filesys.Root = &real.Dir{Path: exposeDir}
	fs.WithCreateFile(real.CreateFile)(filesys)
	fs.WithCreateDir(real.CreateDir)(filesys)
	fs.WithRemoveFile(real.Remove)(filesys)
	fs.IgnorePermissions()(filesys)

	server := filesys.Server()

	errChan := make(chan error)

	// TODO: Use a channel to pass back this if it occurs within a
	// reasonable timeframe.
	go func() {
		errChan <- serveListener(listener, server)
		close(errChan)
	}()

	toReturn := new(Server)
	toReturn.listener = listener
	toReturn.server = &server
	toReturn.errChan = errChan

	// Just before returning, check to see if we got an error off server
	// startup.
	select {
	case err := <-errChan:
		return nil, fmt.Errorf("starting 9p server: %w", err)
	default:
		logrus.Infof("Successfully started 9p server for directory %s", exposeDir)
	}

	return toReturn, nil
}

// Stop a running server.
// Please note that this does *BAD THINGS* to clients if they are still running
// when the server stops. Processes get stuck in I/O deep sleep and zombify, and
// nothing I do save restarting the VM can remove the zombies.
func (s *Server) Stop() error {
	if s.server != nil {
		if err := s.listener.Close(); err != nil {
			return err
		}
		s.server = nil
	}

	return nil
}

// Wait for an error from a running server.
func (s *Server) WaitForError() error {
	if s.server != nil {
		err := <-s.errChan
		return err
	}

	// Server already down, return nil
	return nil
}
