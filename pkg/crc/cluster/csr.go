package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	crcerrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/oc"
	k8scerts "k8s.io/api/certificates/v1beta1"
)

func WaitForOpenshiftResource(ctx context.Context, ocConfig oc.Config, resource string) error {
	logging.Debugf("Waiting for availability of resource type '%s'", resource)
	waitForAPIServer := func() error {
		stdout, stderr, err := ocConfig.WithFailFast().RunOcCommand("get", resource)
		if err != nil {
			logging.Debug(stderr)
			return &crcerrors.RetriableError{Err: err}
		}
		logging.Debug(stdout)
		return nil
	}
	return crcerrors.Retry(ctx, 80*time.Second, waitForAPIServer, time.Second)
}

func getCSRList(ctx context.Context, ocConfig oc.Config, expectedSignerName string) (*k8scerts.CertificateSigningRequestList, error) {
	var csrs k8scerts.CertificateSigningRequestList
	if err := WaitForOpenshiftResource(ctx, ocConfig, "csr"); err != nil {
		return nil, err
	}
	output, stderr, err := ocConfig.WithFailFast().RunOcCommand("get", "csr", "-ojson")
	if err != nil {
		return nil, fmt.Errorf("failed to get all certificate signing requests: %v %s", err, stderr)
	}
	err = json.Unmarshal([]byte(output), &csrs)
	if err != nil {
		return nil, err
	}
	if expectedSignerName == "" {
		return &csrs, nil
	}

	var filteredCsrs []k8scerts.CertificateSigningRequest
	for _, csr := range csrs.Items {
		var signerName string
		if csr.Spec.SignerName != nil {
			signerName = *csr.Spec.SignerName
		}
		if expectedSignerName != signerName {
			continue
		}
		filteredCsrs = append(filteredCsrs, csr)
	}
	csrs.Items = filteredCsrs

	return &csrs, nil
}
