package oc

import (
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

const (
	defaultTimeout = "30s"
	fastTimeout    = "5s"
)

type Config struct {
	Runner           crcos.CommandRunner
	OcExecutablePath string
	KubeconfigPath   string
	Context          string
	Cluster          string
	Timeout          string
}

func (oc Config) WithFailFast() Config {
	return Config{
		Runner:           oc.Runner,
		OcExecutablePath: oc.OcExecutablePath,
		KubeconfigPath:   oc.KubeconfigPath,
		Context:          oc.Context,
		Cluster:          oc.Cluster,
		Timeout:          fastTimeout,
	}
}

func (oc Config) runCommand(isPrivate bool, args ...string) (string, string, error) {
	if oc.Context != "" {
		args = append(args, "--context", oc.Context)
	}
	if oc.Cluster != "" {
		args = append(args, "--cluster", oc.Cluster)
	}
	if oc.KubeconfigPath != "" {
		args = append(args, "--kubeconfig", oc.KubeconfigPath)
	}

	if isPrivate {
		return oc.Runner.RunPrivate("timeout", append([]string{oc.Timeout, oc.OcExecutablePath}, args...)...)
	}

	return oc.Runner.Run("timeout", append([]string{oc.Timeout, oc.OcExecutablePath}, args...)...)
}

func (oc Config) RunOcCommand(args ...string) (string, string, error) {
	return oc.runCommand(false, args...)
}

func (oc Config) RunOcCommandPrivate(args ...string) (string, string, error) {
	return oc.runCommand(true, args...)
}

type SSHRunner struct {
	Runner *ssh.Runner
}

func UseOCWithSSH(sshRunner *ssh.Runner, preset crcPreset.Preset) Config {
	var context, clusterName string
	switch preset {
	case crcPreset.Microshift:
		context = constants.DefaultMicroShiftContext
		clusterName = constants.DefaultMicroShiftClusterName
	case crcPreset.OpenShift:
		context = constants.DefaultOCPContext
		clusterName = constants.DefaultOCPClusterName
	}

	return Config{
		Runner:           sshRunner,
		OcExecutablePath: "oc",
		KubeconfigPath:   "/opt/kubeconfig",
		Context:          context,
		Cluster:          clusterName,
		Timeout:          defaultTimeout,
	}
}
