package machine

import (
	"context"

	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/services/dns"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/pkg/errors"
)

type Preset interface {
	ValidateStartConfig(startConfig types.StartConfig) error
	PreCreate(startConfig types.StartConfig) error
	PreStart(client *client) error
	PostStart(
		ctx context.Context,
		client *client,
		startConfig types.StartConfig,
		servicePostStartConfig dns.ServicePostStartConfig,
		crcBundleMetadata *bundle.CrcBundleInfo,
		instanceIP string,
		sshRunner *crcssh.Runner,
		proxyConfig *network.ProxyConfig,
	) error
	GetOpenShiftStatus(ctx context.Context, ip string, bundle *bundle.CrcBundleInfo) types.OpenshiftStatus
}

type OpenShiftPreset struct {
	Monitoring bool
}

func (p *OpenShiftPreset) PreCreate(startConfig types.StartConfig) error {
	// Ask early for pull secret if it hasn't been requested yet
	_, err := startConfig.PullSecret.Value()
	if err != nil {
		return errors.Wrap(err, "Failed to ask for pull secret")
	}
	return nil
}

func (p *OpenShiftPreset) PreStart(client *client) error {
	if client.useVSock() {
		return exposePorts(openshiftPorts())
	}
	return nil
}

type PodmanPreset struct {
}

func (p *PodmanPreset) ValidateStartConfig(startConfig types.StartConfig) error {
	return nil
}

func (p *PodmanPreset) PreCreate(startConfig types.StartConfig) error {
	return nil
}

func (p *PodmanPreset) PreStart(client *client) error {
	if client.useVSock() {
		return exposePorts(basePorts())
	}
	return nil
}

func (p *PodmanPreset) PostStart(
	ctx context.Context,
	client *client,
	startConfig types.StartConfig,
	servicePostStartConfig dns.ServicePostStartConfig,
	crcBundleMetadata *bundle.CrcBundleInfo,
	instanceIP string,
	sshRunner *crcssh.Runner,
	proxyConfig *network.ProxyConfig,
) error {
	if client.useVSock() {
		if err := dns.CreateResolvFileOnInstance(servicePostStartConfig); err != nil {
			return errors.Wrap(err, "Error running post start")
		}
	}

	return nil
}

func (p *PodmanPreset) GetOpenShiftStatus(ctx context.Context, ip string, bundle *bundle.CrcBundleInfo) types.OpenshiftStatus {
	return types.OpenshiftStopped
}
