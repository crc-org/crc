package machine

import (
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/state"
)

type Client interface {
	GetName() string

	Delete() error
	Exists() (bool, error)
	GetConsoleURL() (*ConsoleResult, error)
	IP() (string, error)
	PowerOff() error
	Start(startConfig StartConfig) (*StartResult, error)
	Status() (*ClusterStatusResult, error)
	Stop() (state.State, error)
	IsRunning() (bool, error)
}

type client struct {
	name              string
	networkMode       network.Mode
	monitoringEnabled bool
}

func NewClient(name string, networkMode network.Mode, monitoringEnabled bool) Client {
	return &client{
		name:              name,
		networkMode:       networkMode,
		monitoringEnabled: monitoringEnabled,
	}
}

func (client *client) GetName() string {
	return client.name
}

func (client *client) useVSock() bool {
	return client.networkMode == network.VSockMode
}
