package services

import (
	"github.com/crc-org/crc/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/pkg/crc/network"
	"github.com/crc-org/crc/pkg/crc/ssh"
)

type ServicePostStartConfig struct {
	Name           string
	SSHRunner      *ssh.Runner
	BundleMetadata bundle.CrcBundleInfo
	IP             string
	NetworkMode    network.Mode
}
