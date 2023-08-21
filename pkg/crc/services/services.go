package services

import (
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
)

type ServicePostStartConfig struct {
	Name           string
	SSHRunner      *ssh.Runner
	BundleMetadata bundle.CrcBundleInfo
	IP             string
	NetworkMode    network.Mode
}
