package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

func WaitForOpenshiftResource(ocConfig oc.Config, resource string) error {
	logging.Debugf("Waiting for availability of resource type '%s'", resource)
	waitForAPIServer := func() error {
		stdout, stderr, err := ocConfig.RunOcCommand("get", resource)
		if err != nil {
			logging.Debug(stderr)
			return &errors.RetriableError{Err: err}
		}
		logging.Debug(stdout)
		return nil
	}
	return errors.RetryAfter(80*time.Second, waitForAPIServer, time.Second)
}

// ApproveNodeCSR approves the certificate for the node.
func ApproveNodeCSR(ocConfig oc.Config) error {
	err := WaitForOpenshiftResource(ocConfig, "csr")
	if err != nil {
		return err
	}

	logging.Debug("Approving pending CSRs")
	// Execute 'oc get csr -oname' and store the output
	csrsJSON, stderr, err := ocConfig.RunOcCommandPrivate("get", "csr", "-ojson")
	if err != nil {
		return fmt.Errorf("Not able to get csr names (%v : %s)", err, stderr)
	}
	var csrs K8sResource
	err = json.Unmarshal([]byte(csrsJSON), &csrs)
	if err != nil {
		return err
	}
	for _, csr := range csrs.Items {
		/* When the CSR hasn't been approved, csr.status is empty in the json data */
		if len(csr.Status.Conditions) == 0 {
			logging.Debugf("Approving csr %s", csr.Metadata.Name)
			_, stderr, err := ocConfig.RunOcCommand("adm", "certificate", "approve", csr.Metadata.Name)
			if err != nil {
				return fmt.Errorf("Not able to approve csr (%v : %s)", err, stderr)
			}
		}
	}
	return nil
}
