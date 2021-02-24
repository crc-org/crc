package cmd

import (
	"net"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/gvisor-tap-vsock/pkg/transport"
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
	return nil, nil
}
