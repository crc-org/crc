package machine

import (
	"context"
	"github.com/pkg/errors"

	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/services/dns"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
)

type Preset interface {
	PreCreate(startConfig types.StartConfig) error
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
}

type OpenShiftLevel2Preset struct {
}

func (p *OpenShiftLevel2Preset) PreCreate(startConfig types.StartConfig) error {
	// Ask early for pull secret if it hasn't been requested yet
	_, err := startConfig.PullSecret.Value()
	if err != nil {
		return errors.Wrap(err, "Failed to ask for pull secret")
	}
	return nil
}

type PodmanPreset struct {
}

func (p *PodmanPreset) PreCreate(startConfig types.StartConfig) error {
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
