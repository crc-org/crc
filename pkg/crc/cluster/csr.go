package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	k8scerts "k8s.io/api/certificates/v1beta1"
)

func WaitForOpenshiftResource(ocConfig oc.Config, resource string) error {
	logging.Debugf("Waiting for availability of resource type '%s'", resource)
	waitForAPIServer := func() error {
		stdout, stderr, err := ocConfig.RunOcCommand("get", resource)
		if err != nil {
			logging.Debug(stderr)
			return &crcerrors.RetriableError{Err: err}
		}
		logging.Debug(stdout)
		return nil
	}
	return crcerrors.RetryAfter(80*time.Second, waitForAPIServer, time.Second)
}

// approveNodeCSR approves the certificate for the node.
func approveNodeCSR(ocConfig oc.Config, expectedSignerName string) error {
	logging.Debug("Approving pending CSRs")
	output, stderr, err := ocConfig.RunOcCommandPrivate("get", "csr", "-ojson")
	if err != nil {
		return fmt.Errorf("Failed to get all certificate signing requests: %v %s", err, stderr)
	}
	var csrs k8scerts.CertificateSigningRequestList
	err = json.Unmarshal([]byte(output), &csrs)
	if err != nil {
		return err
	}
	for _, csr := range csrs.Items {
		/* When the CSR hasn't been approved, csr.status is empty in the json data */
		if len(csr.Status.Conditions) != 0 {
			continue
		}
		var signerName string
		if csr.Spec.SignerName != nil {
			signerName = *csr.Spec.SignerName
		}
		if expectedSignerName != signerName {
			logging.Debugf("Unexpected unapproved csr %s (signerName: %s)", csr.ObjectMeta.Name, signerName)
			continue
		}
		logging.Debugf("Approving csr %s (signerName: %s)", csr.ObjectMeta.Name, signerName)
		_, stderr, err := ocConfig.RunOcCommand("adm", "certificate", "approve", csr.ObjectMeta.Name)
		if err != nil {
			return fmt.Errorf("Not able to approve csr (%v : %s)", err, stderr)
		}
	}
	return nil
}
