package oc

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcos "github.com/code-ready/crc/pkg/os"
)

type OcConfig struct {
	OcBinaryPath   string
	KubeconfigPath string
}

// UseOcWithConfig return the oc binary along with valid kubeconfig
func UseOCWithConfig(machineName string) OcConfig {
	oc := OcConfig{
		OcBinaryPath:   filepath.Join(constants.CrcBinDir, constants.OcBinaryName),
		KubeconfigPath: filepath.Join(constants.MachineInstanceDir, machineName, "kubeconfig"),
	}
	return oc
}

func (oc OcConfig) RunOcCommand(args ...string) (string, string, error) {
	args = append(args, "--kubeconfig", oc.KubeconfigPath)
	return crcos.RunWithDefaultLocale(oc.OcBinaryPath, args...)
}

// ApproveNodeCSR approves the certificate for the node.
func (oc OcConfig) ApproveNodeCSR() error {
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
