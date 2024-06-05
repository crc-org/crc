package daemonclient

import (
	"context"
	"net"
	"net/http"

	"github.com/Microsoft/go-winio"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
)

func transport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return winio.DialPipeContext(ctx, constants.DaemonHTTPNamedPipe)
		},
	}
}
