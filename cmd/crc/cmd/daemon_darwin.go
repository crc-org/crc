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

func setupUnixgramListener() (net.Conn, error) {
	_ = os.Remove(constants.UnixgramSocketPath)
	conn, err := transport.ListenUnixgram(fmt.Sprintf("unixgram://%v", constants.UnixgramSocketPath))
	if err != nil {
		return nil, errors.Wrap(err, "failed to listen unixgram")
	}
	logging.Infof("listening on %s", constants.UnixgramSocketPath)
	vfkitConn, err := transport.AcceptVfkit(conn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to accept vfkit connection")
	}
	return vfkitConn, nil
}

func handleUnixgramConnection(ctx context.Context, vn *virtualnetwork.VirtualNetwork, vfkitConn net.Conn) {
	defer vfkitConn.Close()
	if err := vn.AcceptVfkit(ctx, vfkitConn); err != nil {
		logging.Errorf("failed to accept vfkit connection: %v", err)
	}
	logging.Debugf("Closed connection from %s", vfkitConn.LocalAddr().String())
}

func checkIfDaemonIsRunning() (bool, error) {
	return checkDaemonVersion()
}

func startupDone() {
}
