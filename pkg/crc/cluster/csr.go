package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/pkg/errors"
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

func deleteCSR(ctx context.Context, ocConfig oc.Config, expectedSignerName string) error {
	csrs, err := getCSRList(ctx, ocConfig, expectedSignerName)
	if err != nil {
		return err
	}
	for _, csr := range csrs.Items {
		logging.Debugf("Deleting csr %s (signerName: %s)", csr.ObjectMeta.Name, expectedSignerName)
		_, stderr, err := ocConfig.RunOcCommand("delete", "csr", csr.ObjectMeta.Name)
		if err != nil {
			return errors.Wrapf(err, stderr, "Not able to delete csr (%v : %s)")
		}
	}
	return nil
}

func getCSRList(ctx context.Context, ocConfig oc.Config, expectedSignerName string) (*k8scerts.CertificateSigningRequestList, error) {
	var csrs k8scerts.CertificateSigningRequestList
	if err := WaitForOpenshiftResource(ctx, ocConfig, "csr"); err != nil {
		return nil, err
	}
	output, stderr, err := ocConfig.WithFailFast().RunOcCommand("get", "csr", "-ojson")
	if err != nil {
		return nil, fmt.Errorf("Failed to get all certificate signing requests: %v %s", err, stderr)
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
