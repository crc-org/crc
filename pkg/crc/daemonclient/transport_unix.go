// +build !windows

package daemonclient

import (
	"context"
	"net"
	"net/http"

	"github.com/code-ready/crc/pkg/crc/constants"
)

func transport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", constants.DaemonHTTPSocketPath)
		},
	}
}
