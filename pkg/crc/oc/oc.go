package oc

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/ssh"
	crcos "github.com/code-ready/crc/pkg/os"
)

type Runner interface {
	Run(args ...string) (string, string, error)
	RunPrivate(args ...string) (string, string, error)
}

type Config struct {
	Runner  Runner
	Context string
	Cluster string
}

type LocalRunner struct {
	OcBinaryPath   string
	KubeconfigPath string
}

func (oc LocalRunner) Run(args ...string) (string, string, error) {
	args = append(args, "--kubeconfig", oc.KubeconfigPath)
	return crcos.RunWithDefaultLocale(oc.OcBinaryPath, args...)
}

func (oc LocalRunner) RunPrivate(args ...string) (string, string, error) {
	args = append(args, "--kubeconfig", oc.KubeconfigPath)
	return crcos.RunWithDefaultLocalePrivate(oc.OcBinaryPath, args...)
}

type EnvRunner struct {
	OcBinaryPath  string
	KubeconfigEnv string
}

func (oc EnvRunner) Run(args ...string) (string, string, error) {
	return crcos.RunWithDefaultLocaleAndEnv(oc.OcBinaryPath, args, map[string]string{
		"KUBECONFIG": oc.KubeconfigEnv,
	})
}

func (oc EnvRunner) RunPrivate(args ...string) (string, string, error) {
	return crcos.RunWithDefaultLocalePrivateAndEnv(oc.OcBinaryPath, args, map[string]string{
		"KUBECONFIG": oc.KubeconfigEnv,
	})
}

// UseOcWithConfig return the oc binary along with valid kubeconfig
func UseOCWithConfig(machineName string) Config {
	localRunner := LocalRunner{
		OcBinaryPath:   filepath.Join(constants.CrcOcBinDir, constants.OcBinaryName),
		KubeconfigPath: filepath.Join(constants.MachineInstanceDir, machineName, "kubeconfig"),
	}
	return Config{
		Runner:  localRunner,
		Context: constants.DefaultContext,
		Cluster: constants.DefaultName,
	}
}

func (oc Config) RunOcCommand(args ...string) (string, string, error) {
	if oc.Context != "" {
		args = append(args, "--context", oc.Context)
	}
	if oc.Cluster != "" {
		args = append(args, "--cluster", oc.Cluster)
	}
	return oc.Runner.Run(args...)
}

func (oc Config) RunOcCommandPrivate(args ...string) (string, string, error) {
	if oc.Context != "" {
		args = append(args, "--context", oc.Context)
	}
	if oc.Cluster != "" {
		args = append(args, "--cluster", oc.Cluster)
	}
	return oc.Runner.RunPrivate(args...)
}

type SSHRunner struct {
	Runner *ssh.Runner
}

func UseOCWithSSH(sshRunner *ssh.Runner) Config {
	return Config{
		Runner: SSHRunner{
			Runner: sshRunner,
		},
		Context: constants.DefaultContext,
		Cluster: constants.DefaultName,
	}
}

func (oc SSHRunner) Run(args ...string) (string, string, error) {
	command := fmt.Sprintf("oc --kubeconfig /opt/kubeconfig %s", strings.Join(args, " "))
	stdout, err := oc.Runner.Run(command)
	return stdout, "", err
}

func (oc SSHRunner) RunPrivate(args ...string) (string, string, error) {
	command := fmt.Sprintf("oc --kubeconfig /opt/kubeconfig %s", strings.Join(args, " "))
	stdout, err := oc.Runner.RunPrivate(command)
	return stdout, "", err
}
