package daemonclient

import (
	"net/http"

	networkclient "github.com/code-ready/gvisor-tap-vsock/pkg/client"
)

type Client struct {
	*networkclient.Client
}

func New() *Client {
	return &Client{
		Client: networkclient.New(&http.Client{
			Transport: transport(),
		}, "http://unix/network"),
	}
}
