package adminhelper

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/crc-org/admin-helper/pkg/client"
)

func instance() helper {
	return Client()
}

func Client() *client.Client {
	return client.New(&http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, `\\.\pipe\crc-admin-helper`)
			},
		},
		Timeout: 3 * time.Second,
	}, "http://unix")
}
