//go:build !windows
// +build !windows

package daemonclient

import (
	"context"
	"net"
	"net/http"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
)

func transport() *http.Transport {
	return &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", constants.DaemonHTTPSocketPath)
		},
	}
}
