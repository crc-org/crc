package cmd

import (
	"net"

	"github.com/Microsoft/go-winio"
	"github.com/containers/gvisor-tap-vsock/pkg/transport"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
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
	logging.Infof("listening %s", constants.DaemonHTTPNamedPipe)
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
