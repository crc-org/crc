package daemonclient

import (
	"net/http"

	"github.com/code-ready/crc/pkg/crc/api/client"
	networkclient "github.com/containers/gvisor-tap-vsock/pkg/client"
)

type Client struct {
	NetworkClient *networkclient.Client
	APIClient     *client.Client
}

func New() *Client {
	return &Client{
		NetworkClient: networkclient.New(&http.Client{
			Transport: transport(),
		}, "http://unix/network"),
		APIClient: client.New(&http.Client{
			Transport: transport(),
		}, "http://unix/api"),
	}
}
