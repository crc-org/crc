package daemonclient

import (
	"context"
	"net"
	"net/http"

	"github.com/Microsoft/go-winio"
	"github.com/code-ready/crc/pkg/crc/constants"
)

func transport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return winio.DialPipeContext(ctx, constants.DaemonHTTPNamedPipe)
		},
	}
}
