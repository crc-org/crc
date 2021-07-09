package cmd

import (
	"net"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
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

func startupDone() {
}
