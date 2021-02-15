package fakemachine

import (
	"context"
	"errors"

	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/state"
)

func NewClient() *Client {
	return &Client{}
}

func NewFailingClient() *Client {
	return &Client{
		Failing: true,
	}
}

type Client struct {
	Failing bool
}

var DummyClusterConfig = machine.ClusterConfig{
	ClusterCACert: "MIIDODCCAiCgAwIBAgIIRVfCKNUa1wIwDQYJ",
	KubeConfig:    "/tmp/kubeconfig",
	KubeAdminPass: "foobar",
	ClusterAPI:    "https://foo.testing:6443",
	WebConsoleURL: "https://console.foo.testing:6443",
	ProxyConfig:   nil,
}

func (c *Client) GetName() string {
	return "crc"
}

func (c *Client) Delete() error {
	if c.Failing {
		return errors.New("delete failed")
	}
	return nil
}

func (c *Client) GetConsoleURL() (*machine.ConsoleResult, error) {
	if c.Failing {
		return nil, errors.New("console failed")
	}
	return &machine.ConsoleResult{
		ClusterConfig: DummyClusterConfig,
		State:         state.Running,
	}, nil
}

func (c *Client) GetProxyConfig(machineName string) (*network.ProxyConfig, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) IP() (string, error) {
	return "", errors.New("not implemented")
}

func (c *Client) PowerOff() error {
	if c.Failing {
		return errors.New("poweroff failed")
	}
	return nil
}

func (c *Client) Start(ctx context.Context, startConfig machine.StartConfig) (*machine.StartResult, error) {
	if c.Failing {
		return nil, errors.New("Failed to start")
	}
	return &machine.StartResult{
		ClusterConfig:  DummyClusterConfig,
		KubeletStarted: true,
	}, nil
}

func (c *Client) Stop() (state.State, error) {
	if c.Failing {
		return state.Running, errors.New("stop failed")
	}
	return state.Stopped, nil
}

func (c *Client) Status() (*machine.ClusterStatusResult, error) {
	if c.Failing {
		return nil, errors.New("broken")
	}
	return &machine.ClusterStatusResult{
		CrcStatus:        state.Running,
		OpenshiftStatus:  "Running",
		OpenshiftVersion: "4.5.1",
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
	}, nil
}

func (c *Client) Exists() (bool, error) {
	return true, nil
}

func (c *Client) IsRunning() (bool, error) {
	return true, nil
}
