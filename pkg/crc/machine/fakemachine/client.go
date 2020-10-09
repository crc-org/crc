package fakemachine

import (
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

func (c *Client) Delete(deleteConfig machine.DeleteConfig) error {
	if c.Failing {
		return errors.New("delete failed")
	}
	return nil
}

func (c *Client) GetConsoleURL(consoleConfig machine.ConsoleConfig) (machine.ConsoleResult, error) {
	if c.Failing {
		return machine.ConsoleResult{
			ClusterConfig: DummyClusterConfig,
			Success:       false,
			Error:         "console failed",
			State:         state.Running,
		}, errors.New("console failed")
	}
	return machine.ConsoleResult{
		ClusterConfig: DummyClusterConfig,
		Success:       true,
		State:         state.Running,
	}, nil
}

func (c *Client) GetProxyConfig(machineName string) (*network.ProxyConfig, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) IP(ipConfig machine.IPConfig) (string, error) {
	return "", errors.New("not implemented")
}

func (c *Client) PowerOff(powerOff machine.PowerOffConfig) (machine.PowerOffResult, error) {
	if c.Failing {
		return machine.PowerOffResult{
			Name:    "crc",
			Success: false,
			Error:   "poweroff failed",
		}, errors.New("poweroff failed")
	}
	return machine.PowerOffResult{
		Name:    "crc",
		Success: true,
	}, nil
}

func (c *Client) Start(startConfig machine.StartConfig) (machine.StartResult, error) {
	if c.Failing {
		return machine.StartResult{
			Name:           startConfig.Name,
			Error:          "Failed to start",
			KubeletStarted: false,
		}, nil
	}
	return machine.StartResult{
		Name:           startConfig.Name,
		ClusterConfig:  DummyClusterConfig,
		KubeletStarted: true,
	}, nil
}

func (c *Client) Stop(stopConfig machine.StopConfig) (machine.StopResult, error) {
	if c.Failing {
		return machine.StopResult{
			Name:    "crc",
			Success: false,
			Error:   "stop failed",
			State:   state.Running,
		}, errors.New("stop failed")
	}
	return machine.StopResult{
		Name:    "crc",
		Success: true,
		State:   state.Stopped,
	}, nil
}

func (c *Client) Status(statusConfig machine.ClusterStatusConfig) (machine.ClusterStatusResult, error) {
	if c.Failing {
		return machine.ClusterStatusResult{
			Success: false,
			Error:   "broken",
		}, errors.New("broken")
	}
	return machine.ClusterStatusResult{
		Name:             "crc",
		CrcStatus:        "Running",
		OpenshiftStatus:  "Running",
		OpenshiftVersion: "4.5.1",
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
		Success:          true,
	}, nil
}

func (c *Client) Exists(name string) (bool, error) {
	return true, nil
}
