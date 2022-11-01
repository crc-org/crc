package libmachine

import (
	"io"

	"github.com/crc-org/crc/pkg/libmachine/host"
	"github.com/crc-org/crc/pkg/libmachine/persist"
	rpcdriver "github.com/crc-org/machine/libmachine/drivers/rpc"
)

type API interface {
	io.Closer
	NewHost(driverName string, driverPath string, rawDriver []byte) (*host.Host, error)
	persist.Store
}

type Client struct {
	*persist.Filestore
	clientDriverFactory rpcdriver.RPCClientDriverFactory
}

func NewClient(storePath string) *Client {
	return &Client{
		Filestore:           persist.NewFilestore(storePath),
		clientDriverFactory: rpcdriver.NewRPCClientDriverFactory(),
	}
}

func (api *Client) Close() error {
	return api.clientDriverFactory.Close()
}
