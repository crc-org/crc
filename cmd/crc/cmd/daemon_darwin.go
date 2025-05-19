package cmd

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/containers/gvisor-tap-vsock/pkg/transport"
	"github.com/containers/gvisor-tap-vsock/pkg/virtualnetwork"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/pkg/errors"
)

func vsockListener() (net.Listener, error) {
	_ = os.Remove(constants.TapSocketPath)
	ln, err := net.Listen("unix", constants.TapSocketPath)
	logging.Infof("listening %s", constants.TapSocketPath)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func httpListener() (net.Listener, error) {
	_ = os.Remove(constants.DaemonHTTPSocketPath)
	ln, err := net.Listen("unix", constants.DaemonHTTPSocketPath)
	logging.Infof("listening %s", constants.DaemonHTTPSocketPath)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func unixgramListener(ctx context.Context, vn *virtualnetwork.VirtualNetwork) (*net.UnixConn, error) {
	_ = os.Remove(constants.UnixgramSocketPath)
	conn, err := transport.ListenUnixgram(fmt.Sprintf("unixgram://%v", constants.UnixgramSocketPath))
	if err != nil {
		return conn, errors.Wrap(err, "failed to listen unixgram")
	}
	logging.Infof("listening on %s:", constants.UnixgramSocketPath)
	vfkitConn, err := transport.AcceptVfkit(conn)
	if err != nil {
		return conn, errors.Wrap(err, "failed to accept vfkit connection")
	}
	go func() {
		err := vn.AcceptVfkit(ctx, vfkitConn)
		if err != nil {
			logging.Errorf("failed to accept vfkit connection: %v", err)
			return
		}
	}()
	return conn, err
}

func checkIfDaemonIsRunning() (bool, error) {
	return checkDaemonVersion()
}

func startupDone() {
}
