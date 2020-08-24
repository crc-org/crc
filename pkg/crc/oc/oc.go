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
	Run(binaryPath string, args ...string) (string, string, error)
	RunPrivate(binaryPath string, args ...string) (string, string, error)
}

type Config struct {
	Runner         Runner
	OcBinaryPath   string
	KubeconfigPath string
	Context        string
	Cluster        string
}

type LocalRunner struct {
}

func (oc LocalRunner) Run(binaryPath string, args ...string) (string, string, error) {
	return crcos.RunWithDefaultLocale(binaryPath, args...)
}

func (oc LocalRunner) RunPrivate(binaryPath string, args ...string) (string, string, error) {
	return crcos.RunWithDefaultLocalePrivate(binaryPath, args...)
}

// UseOcWithConfig return the oc binary along with valid kubeconfig
func UseOCWithConfig(machineName string) Config {
	return Config{
		Runner:         LocalRunner{},
		OcBinaryPath:   filepath.Join(constants.CrcOcBinDir, constants.OcBinaryName),
		KubeconfigPath: filepath.Join(constants.MachineInstanceDir, machineName, "kubeconfig"),
		Context:        constants.DefaultContext,
		Cluster:        constants.DefaultName,
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
		return oc.Runner.RunPrivate(oc.OcBinaryPath, args...)
	}

	return oc.Runner.Run(oc.OcBinaryPath, args...)
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
		Runner: SSHRunner{
			Runner: sshRunner,
		},
		OcBinaryPath:   "oc",
		KubeconfigPath: "/opt/kubeconfig",
		Context:        constants.DefaultContext,
		Cluster:        constants.DefaultName,
	}
}

func (oc SSHRunner) Run(binaryPath string, args ...string) (string, string, error) {
	command := fmt.Sprintf("%s %s", binaryPath, strings.Join(args, " "))
	stdout, err := oc.Runner.Run(command)
	return stdout, "", err
}

func (oc SSHRunner) RunPrivate(binaryPath string, args ...string) (string, string, error) {
	command := fmt.Sprintf("%s %s", binaryPath, strings.Join(args, " "))
	stdout, err := oc.Runner.RunPrivate(command)
	return stdout, "", err
}
