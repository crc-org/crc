package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/Microsoft/go-winio"
	"github.com/containers/gvisor-tap-vsock/pkg/transport"
	"github.com/containers/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/fileserver/fs9p"
	"github.com/crc-org/machine/libmachine/drivers"
)

func vsockListener() (net.Listener, error) {
	ln, err := transport.Listen(transport.DefaultURL)
	logging.Infof("listening %s", transport.DefaultURL)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func httpListener() (net.Listener, error) {
	ln, err := winio.ListenPipe(constants.DaemonHTTPNamedPipe, &winio.PipeConfig{
		MessageMode:      true,  // Use message mode so that CloseWrite() is supported
		InputBufferSize:  65536, // Use 64kB buffers to improve performance
		OutputBufferSize: 65536,
	})
	if err != nil {
		return nil, err
	}
	logging.Infof("listening %s", constants.DaemonHTTPNamedPipe)
	return ln, nil
}

func checkIfDaemonIsRunning() (bool, error) {
	return checkDaemonVersion()
}

func unixgramListener(_ context.Context, _ *virtualnetwork.VirtualNetwork) (*net.UnixConn, error) {
	return nil, drivers.ErrNotImplemented
}

func startupDone() {
}

func startSharedDirServers(vn *virtualnetwork.VirtualNetwork, gatewayIP string, enabled bool) (func(), error) {
	noop := func() {}
	if !enabled {
		return noop, nil
	}

	// 9p over hvsock
	listener9pHvsock, err := fs9p.GetHvsockListener(constants.Plan9HvsockGUID)
	if err != nil {
		return noop, err
	}
	server9pHvsock, err := fs9p.New9pServer(listener9pHvsock, constants.GetHomeDir())
	if err != nil {
		return noop, err
	}
	if err := server9pHvsock.Start(); err != nil {
		return noop, err
	}
	cleanup := func() {
		if err := server9pHvsock.Stop(); err != nil {
			logging.Warnf("error stopping 9p server (hvsock): %v", err)
		}
	}
	go func() {
		if err := server9pHvsock.WaitForError(); err != nil {
			logging.Errorf("9p server (hvsock) error: %v", err)
		}
	}()

	// 9p over TCP (as a backup)
	listener9pTCP, err := vn.Listen("tcp", net.JoinHostPort(gatewayIP, fmt.Sprintf("%d", constants.Plan9TcpPort)))
	if err != nil {
		cleanup()
		return noop, err
	}
	server9pTCP, err := fs9p.New9pServer(listener9pTCP, constants.GetHomeDir())
	if err != nil {
		cleanup()
		return noop, err
	}
	if err := server9pTCP.Start(); err != nil {
		cleanup()
		return noop, err
	}
	cleanup = func() {
		if err := server9pHvsock.Stop(); err != nil {
			logging.Warnf("error stopping 9p server (hvsock): %v", err)
		}
		if err := server9pTCP.Stop(); err != nil {
			logging.Warnf("error stopping 9p server (tcp): %v", err)
		}
	}
	go func() {
		if err := server9pTCP.WaitForError(); err != nil {
			logging.Errorf("9p server (tcp) error: %v", err)
		}
	}()

	return cleanup, nil
}
