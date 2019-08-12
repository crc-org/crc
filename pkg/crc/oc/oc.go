package oc

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
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
	certNamses, stderr, err := oc.RunOcCommand("get", "csr", "-oname")
	if err != nil {
		return fmt.Errorf("Not able to get csr names (%v : %s)", err, stderr)
	}

	// Split the output with new line and run 'oc adm certificate approve <certName>'
	for _, certName := range strings.Split(certNamses, "\n") {
		if certName == "" {
			continue
		}
		_, stderr, err := oc.RunOcCommand("adm", "certificate", "approve", certName)
		if err != nil {
			return fmt.Errorf("Not able to get csr names (%v : %s)", err, stderr)
		}
	}
	return nil
}
