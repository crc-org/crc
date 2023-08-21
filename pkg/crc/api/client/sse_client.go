package client

import (
	"encoding/json"
	"net/http"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/r3labs/sse/v2"
)

type SSEClient struct {
	client *sse.Client
}

func NewSSEClient(transport *http.Transport) *SSEClient {
	client := sse.NewClient("http://unix/events")
	client.Connection.Transport = transport
	return &SSEClient{
		client: client,
	}
}

func (c *SSEClient) Status(statusCallback func(*types.ClusterLoadResult)) error {
	err := c.client.Subscribe("status", func(msg *sse.Event) {
		wmState := &types.ClusterLoadResult{}
		err := json.Unmarshal(msg.Data, wmState)
		if err != nil {
			logging.Errorf("Could not parse status event: %s", err)
			return
		}
		statusCallback(wmState)
	})

	return err
}
