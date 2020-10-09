package machine

import "github.com/code-ready/machine/libmachine/state"

type Client interface {
	Delete(deleteConfig DeleteConfig) error
	Exists(name string) (bool, error)
	GetConsoleURL(consoleConfig ConsoleConfig) (*ConsoleResult, error)
	IP(ipConfig IPConfig) (string, error)
	PowerOff(powerOff PowerOffConfig) error
	Start(startConfig StartConfig) (*StartResult, error)
	Status(statusConfig ClusterStatusConfig) (*ClusterStatusResult, error)
	Stop(stopConfig StopConfig) (state.State, error)
}

type client struct {
	debug bool
}

func NewClient(debug bool) Client {
	return &client{
		debug: debug,
	}
}
