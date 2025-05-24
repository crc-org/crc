package fs9p

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/docker/go-p9p"
	"github.com/docker/go-p9p/ufs"
	"github.com/sirupsen/logrus"
)

type Server struct {
	listener net.Listener
	// Errors from the server being started will come out here.
	errChan chan error
}

func Serve9p(listener net.Listener, exposeDir string) error {
	defer listener.Close()

	ctx := context.Background()

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Fatalln("error accepting:", err)
		}

		go func(conn net.Conn) {
			defer conn.Close()

			log.Println("connected", conn.RemoteAddr())

			ctx := context.WithValue(ctx, "conn", conn)
			session, err := ufs.NewSession(ctx, exposeDir)
			if err != nil {
				log.Println("error creating session")
				return
			}

			if err := p9p.ServeConn(ctx, conn, p9p.Dispatch(session)); err != nil {
				log.Printf("serving conn: %v", err)
			}
		}(c)
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

	errChan := make(chan error)

	go func() {
		errChan <- Serve9p(listener, exposeDir)
		close(errChan)
	}()

	toReturn := new(Server)
	toReturn.listener = listener
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
	//if s.server != nil {
	if err := s.listener.Close(); err != nil {
		return err
	}
	//s.server = nil
	//}

	return nil
}

// Wait for an error from a running server.
func (s *Server) WaitForError() error {
	//if s.server != nil {
	err := <-s.errChan
	return err
	//}

	// Server already down, return nil
	//return nil
}
