// +build !windows

package api

import (
	"net"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func createIPCListener(socketPath string) (net.Listener, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logging.Errorf("Failed to create socket at %s: %s", socketPath, err.Error())
		return nil, err
	}
	return listener, nil
}
