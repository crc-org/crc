package oc

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/ssh"
	crcos "github.com/code-ready/crc/pkg/os"
)

type Config struct {
	Runner           crcos.CommandRunner
	OcExecutablePath string
	KubeconfigPath   string
	Context          string
	Cluster          string
}

// UseOcWithConfig return the oc executable along with valid kubeconfig
func UseOCWithConfig(machineName string) Config {
	return Config{
		Runner:           crcos.NewLocalCommandRunner(),
		OcExecutablePath: filepath.Join(constants.CrcOcBinDir, constants.OcExecutableName),
		KubeconfigPath:   filepath.Join(constants.MachineInstanceDir, machineName, "kubeconfig"),
		Context:          constants.DefaultContext,
		Cluster:          constants.DefaultName,
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
		return oc.Runner.RunPrivate(oc.OcExecutablePath, args...)
	}

	return oc.Runner.Run(oc.OcExecutablePath, args...)
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

func UseOCWithSSH(sshRunner *ssh.Runner) Config {
	return Config{
		Runner:           sshRunner,
		OcExecutablePath: "oc",
		KubeconfigPath:   "/opt/kubeconfig",
		Context:          constants.DefaultContext,
		Cluster:          constants.DefaultName,
	}
}
