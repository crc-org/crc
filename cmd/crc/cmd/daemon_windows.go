package cmd

import (
	"net"

	"github.com/Microsoft/go-winio"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
	"github.com/containers/gvisor-tap-vsock/pkg/transport"
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

func daemonNotRunningMessage() string {
	if crcversion.IsInstaller() {
		return "Is CodeReady Containers tray application running? Cannot reach daemon API"
	}
	return genericDaemonNotRunningMessage
}

func startupDone() {
}
