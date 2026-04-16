package cmd

import (
	"net"
	"os"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

func httpListener() (net.Listener, error) {
	_ = os.Remove(constants.DaemonHTTPSocketPath)
	ln, err := net.Listen("unix", constants.DaemonHTTPSocketPath)
	logging.Infof("listening %s", constants.DaemonHTTPSocketPath)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func adminHelperListener() (net.Listener, error) {
	addr := "127.0.0.1:9764"
	ln, err := net.Listen("tcp", addr)
	logging.Infof("admin helper listening %s", addr)
	if err != nil {
		return nil, err
	}
	return ln, nil
}

func checkIfDaemonIsRunning() (bool, error) {
	return checkDaemonVersion()
}

func startupDone() {
}
