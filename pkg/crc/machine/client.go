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
}

type client struct {
	name        string
	debug       bool
	networkMode network.Mode
}

func NewClient(name string, debug bool, networkMode network.Mode) Client {
	return &client{
		name:        name,
		debug:       debug,
		networkMode: networkMode,
	}
}

func (client *client) GetName() string {
	return client.name
}

func (client *client) useVSock() bool {
	return client.networkMode == network.VSockMode
}
