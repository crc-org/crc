package adminhelper

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/code-ready/admin-helper/pkg/client"
)

func instance() helper {
	return Client()
}

func Client() *client.Client {
	return client.New(&http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, `\\.\pipe\crc-admin-helper`)
			},
		},
		Timeout: 3 * time.Second,
	}, "http://unix")
}
