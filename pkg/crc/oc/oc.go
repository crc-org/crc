package oc

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

type OcRunner interface {
	Run(args ...string) (string, string, error)
	RunPrivate(args ...string) (string, string, error)
}

type OcConfig struct {
	Runner  OcRunner
	Context string
	Cluster string
}

type OcLocalRunner struct {
	OcBinaryPath   string
	KubeconfigPath string
}

func (oc OcLocalRunner) Run(args ...string) (string, string, error) {
	args = append(args, "--kubeconfig", oc.KubeconfigPath)
	return crcos.RunWithDefaultLocale(oc.OcBinaryPath, args...)
}

func (oc OcLocalRunner) RunPrivate(args ...string) (string, string, error) {
	args = append(args, "--kubeconfig", oc.KubeconfigPath)
	return crcos.RunWithDefaultLocalePrivate(oc.OcBinaryPath, args...)
}

// UseOcWithConfig return the oc binary along with valid kubeconfig
func UseOCWithConfig(machineName string) OcConfig {
	localRunner := OcLocalRunner{
		OcBinaryPath:   filepath.Join(constants.CrcOcBinDir, constants.OcBinaryName),
		KubeconfigPath: filepath.Join(constants.MachineInstanceDir, machineName, "kubeconfig"),
	}
	return NewOcConfig(localRunner, constants.DefaultContext, constants.DefaultName)
}

func (oc OcConfig) RunOcCommand(args ...string) (string, string, error) {
	args = append(args, "--context", oc.Context, "--cluster", oc.Cluster)
	return oc.Runner.Run(args...)
}

func (oc OcConfig) RunOcCommandPrivate(args ...string) (string, string, error) {
	args = append(args, "--context", oc.Context, "--cluster", oc.Cluster)
	return oc.Runner.RunPrivate(args...)
}

func NewOcConfig(runner OcRunner, context string, clusterName string) OcConfig {
	return OcConfig{
		Runner:  runner,
		Context: context,
		Cluster: clusterName,
	}
}
