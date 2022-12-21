package client

import (
	"context"
	"net/http"

	"github.com/crc-org/crc/pkg/crc/machine/types"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WebSocketClient struct {
	client           *http.Client
	statusConnection *websocket.Conn
}

func NewWebSocketClient(httpClient *http.Client) *WebSocketClient {
	return &WebSocketClient{
		client: httpClient,
	}
}

func (c *WebSocketClient) Status() (*types.ClusterLoadResult, error) {
	ctx := context.Background()

	if c.statusConnection == nil {
		conn, _, err := websocket.Dial(ctx, "ws://unix/socket/status", &websocket.DialOptions{HTTPClient: c.client})

		if err != nil {
			return nil, err
		}
		c.statusConnection = conn
	}

	wmState := &types.ClusterLoadResult{}
	err := wsjson.Read(ctx, c.statusConnection, wmState)
	return wmState, err

}
