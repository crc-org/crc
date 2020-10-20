package services

import (
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/ssh"
)

type ServicePostStartConfig struct {
	Name           string
	SSHRunner      *ssh.Runner
	BundleMetadata bundle.CrcBundleInfo
	IP             string
}
