package fakemachine

import (
	"context"
	"errors"

	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
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

var DummyClusterConfig = types.ClusterConfig{
	ClusterType:   "openshift",
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

func (c *Client) GetConsoleURL() (*types.ConsoleResult, error) {
	if c.Failing {
		return nil, errors.New("console failed")
	}
	return &types.ConsoleResult{
		ClusterConfig: DummyClusterConfig,
		State:         state.Running,
	}, nil
}

func (c *Client) GetProxyConfig(_ string) (*httpproxy.ProxyConfig, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) ConnectionDetails() (*types.ConnectionDetails, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) PowerOff() error {
	if c.Failing {
		return errors.New("poweroff failed")
	}
	return nil
}

func (c *Client) GenerateBundle(_ bool) error {
	if c.Failing {
		return errors.New("bundle generation failed")
	}
	return nil
}

func (c *Client) Start(_ context.Context, _ types.StartConfig) (*types.StartResult, error) {
	if c.Failing {
		return nil, errors.New("Failed to start")
	}
	return &types.StartResult{
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

func (c *Client) Status() (*types.ClusterStatusResult, error) {
	if c.Failing {
		return nil, errors.New("broken")
	}
	return &types.ClusterStatusResult{
		CrcStatus:        state.Running,
		OpenshiftStatus:  types.OpenshiftRunning,
		OpenshiftVersion: "4.5.1",
		PodmanVersion:    "3.3.1",
		DiskUse:          10_000_000_000,
		DiskSize:         20_000_000_000,
		RAMSize:          2_000,
		RAMUse:           1_000,
		Preset:           preset.OpenShift,
	}, nil
}

func (c *Client) Exists() (bool, error) {
	return true, nil
}

func (c *Client) IsRunning() (bool, error) {
	return true, nil
}

func (c *Client) GetPreset() preset.Preset {
	return preset.OpenShift
}

func (c *Client) GetClusterLoad() (*types.ClusterLoadResult, error) {
	return nil, errors.New("not implemented")
}
