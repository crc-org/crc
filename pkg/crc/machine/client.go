package machine

import (
	"github.com/code-ready/machine/libmachine/state"
)

type Client interface {
	Delete(name string) error
	Exists(name string) (bool, error)
	GetConsoleURL(name string) (*ConsoleResult, error)
	IP(name string) (string, error)
	PowerOff(name string) error
	Start(startConfig StartConfig) (*StartResult, error)
	Status(name string) (*ClusterStatusResult, error)
	Stop(name string) (state.State, error)
}

type client struct {
	debug bool
}

func NewClient(debug bool) Client {
	return &client{
		debug: debug,
	}
}
