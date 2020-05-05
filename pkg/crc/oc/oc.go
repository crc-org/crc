package oc

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
)

type OcRunner interface {
	Run(args ...string) (string, string, error)
	RunPrivate(args ...string) (string, string, error)
	GetKubeconfigPath() string
}

type OcConfig struct {
	runner  OcRunner
	Context string
	Cluster string
}

type OcLocalRunner struct {
	OcBinaryPath   string
	KubeconfigPath string
}

func (oc OcLocalRunner) Run(args ...string) (string, string, error) {
	return crcos.RunWithDefaultLocale(oc.OcBinaryPath, args...)
}

func (oc OcLocalRunner) RunPrivate(args ...string) (string, string, error) {
	return crcos.RunWithDefaultLocalePrivate(oc.OcBinaryPath, args...)
}

func (oc OcLocalRunner) GetKubeconfigPath() string {
	return oc.KubeconfigPath
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
	args = append(args, "--kubeconfig", oc.runner.GetKubeconfigPath(), "--context", oc.Context, "--cluster", oc.Cluster)
	return oc.runner.Run(args...)
}

func (oc OcConfig) RunOcCommandPrivate(args ...string) (string, string, error) {
	args = append(args, "--kubeconfig", oc.runner.GetKubeconfigPath(), "--context", oc.Context, "--cluster", oc.Cluster)
	return oc.runner.RunPrivate(args...)
}

func NewOcConfig(runner OcRunner, context string, clusterName string) OcConfig {
	return OcConfig{
		runner:  runner,
		Context: context,
		Cluster: clusterName,
	}
}

func (oc OcConfig) WaitForOpenshiftResource(resource string) error {
	waitForApiServer := func() error {
		stdout, stderr, err := oc.RunOcCommand("get", resource)
		if err != nil {
			logging.Debug(stderr)
			return &errors.RetriableError{Err: err}
		}
		logging.Debug(stdout)
		return nil
	}
	return errors.RetryAfter(80, waitForApiServer, time.Second)
}

// ApproveNodeCSR approves the certificate for the node.
func (oc OcConfig) ApproveNodeCSR() error {
	err := oc.WaitForOpenshiftResource("csr")
	if err != nil {
		return err
	}
	// Execute 'oc get csr -oname' and store the output
	csrsJson, stderr, err := oc.RunOcCommand("get", "csr", "-ojson")
	if err != nil {
		return fmt.Errorf("Not able to get csr names (%v : %s)", err, stderr)
	}
	var csrs K8sResource
	err = json.Unmarshal([]byte(csrsJson), &csrs)
	if err != nil {
		return err
	}
	for _, csr := range csrs.Items {
		/* When the CSR hasn't been approved, csr.status is empty in the json data */
		if len(csr.Status.Conditions) == 0 {
			logging.Debugf("Approving csr %s", csr.Metadata.Name)
			_, stderr, err := oc.RunOcCommand("adm", "certificate", "approve", csr.Metadata.Name)
			if err != nil {
				return fmt.Errorf("Not able to approve csr (%v : %s)", err, stderr)
			}
		}
	}
	return nil
}
