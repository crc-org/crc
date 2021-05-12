package dns

import (
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/ssh"
)

type ServicePostStartConfig struct {
	Name           string
	SSHRunner      *ssh.Runner
	BundleMetadata bundle.CrcBundleInfo
	IP             string
	NetworkMode    network.Mode
}
