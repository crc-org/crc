package machine

import (
	"context"
	"time"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/kofalt/go-memoize"
)

type Client interface {
	GetName() string
	GetConsoleURL() (*types.ConsoleResult, error)
	ConnectionDetails() (*types.ConnectionDetails, error)

	Delete() error
	Exists() (bool, error)
	PowerOff() error
	Start(ctx context.Context, startConfig types.StartConfig) (*types.StartResult, error)
	Status() (*types.ClusterStatusResult, error)
	Stop() (state.State, error)
	IsRunning() (bool, error)
	GenerateBundle(forceStop bool) error
}

type client struct {
	name   string
	debug  bool
	config crcConfig.Storage

	diskDetails *memoize.Memoizer
}

func NewClient(name string, debug bool, config crcConfig.Storage) Client {
	return &client{
		name:        name,
		debug:       debug,
		config:      config,
		diskDetails: memoize.NewMemoizer(time.Minute, 5*time.Minute),
	}
}

func (client *client) GetName() string {
	return client.name
}

func (client *client) useVSock() bool {
	return client.networkMode() == network.UserNetworkingMode
}

func (client *client) networkMode() network.Mode {
	return crcConfig.GetNetworkMode(client.config)
}

func (client *client) monitoringEnabled() bool {
	return client.config.Get(crcConfig.EnableClusterMonitoring).AsBool()
}
