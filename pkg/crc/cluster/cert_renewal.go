package cluster

import (
	"context"
	"fmt"
	"time"

	crcerrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/oc"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"

	k8scerts "k8s.io/api/certificates/v1beta1"
)

func isPending(csr *k8scerts.CertificateSigningRequest) bool {
	return len(csr.Status.Conditions) == 0 && len(csr.Status.Certificate) == 0
}

func approvePendingCSRs(ctx context.Context, ocConfig oc.Config, expectedSignerName string) error {
	return crcerrors.Retry(ctx, 10*time.Minute, func() error {
		csrs, err := getCSRList(ctx, ocConfig, expectedSignerName)
		if err != nil {
			return &crcerrors.RetriableError{Err: err}
		}
		csrsApproved := false
		for i := range csrs.Items {
			csr := &csrs.Items[i]
			if !isPending(csr) {
				continue
			}
			logging.Debugf("Approving csr %s (signerName: %s)", csr.Name, expectedSignerName)
			_, stderr, err := ocConfig.RunOcCommand("adm", "certificate", "approve", csr.Name)
			if err != nil {
				return fmt.Errorf("not able to approve csr (%v : %s)", err, stderr)
			}
			csrsApproved = true
		}
		if !csrsApproved {
			return &crcerrors.RetriableError{Err: fmt.Errorf("no Pending CSR with signerName %s", expectedSignerName)}
		}
		return nil
	}, time.Second*5)
}

func ApproveCSRAndWaitForCertsRenewal(ctx context.Context, sshRunner *ssh.Runner, ocConfig oc.Config, client, server, aggregratorClient bool) error {
	const (
		kubeletClientSignerName  = "kubernetes.io/kube-apiserver-client-kubelet"
		kubeletServingSignerName = "kubernetes.io/kubelet-serving"
	)

	// First, kubelet starts and tries to connect to API server. If its certificate is expired, it asks for a new one
	// Admin needs to approve it. The Kubernetes controller manager will then issue the cert, kubelet will fetch it and use it.
	// Kubelet stores the cert in /var/lib/kubelet/pki/kubelet-client-current.pem
	if client {
		logging.Info("Kubelet client certificate has expired, renewing it... [will take up to 10 minutes]")
		if err := approvePendingCSRs(ctx, ocConfig, kubeletClientSignerName); err != nil {
			logging.Debugf("Error approving pending kube-apiserver-client-kubelet CSRs: %v", err)
			return err
		}

		if err := approvePendingCSRs(ctx, ocConfig, kubeletServingSignerName); err != nil {
			logging.Debugf("Error approving pending kubelet-serving CSRs: %v", err)
			return err
		}

		if err := crcerrors.Retry(ctx, 5*time.Minute, waitForCertRenewal(sshRunner, KubeletClientCert), time.Second*5); err != nil {
			logging.Debugf("Error approving pending kube-apiserver-client-kubelet CSR: %v", err)
			return err
		}
	}
	// API server needs to connect to kubelet for some features like logs, port forwards. This communication is backed by a cert
	// store in /var/lib/kubelet/pki/kubelet-server-current.pem
	// After kubelet connected to the API server, if the serving cert is expireed, kubelet asks for a new CSR.
	// This CSR is automatically approved by the cluster-machine-approver. The k8s controller manager issues the cert and kubelet fetches it.
	if server {
		logging.Info("Kubelet serving certificate has expired, waiting for automatic renewal... [will take up to 5 minutes]")
		return crcerrors.Retry(ctx, 5*time.Minute, waitForCertRenewal(sshRunner, KubeletServerCert), time.Second*5)
	}
	if aggregratorClient {
		logging.Info("Kube API server certificate has expired, waiting for automatic renewal... [will take up to 8 minutes]")
		return crcerrors.Retry(ctx, 8*time.Minute, waitForCertRenewal(sshRunner, AggregatorClientCert), time.Second*5)
	}
	return nil
}

func waitForCertRenewal(sshRunner *ssh.Runner, cert string) func() error {
	return func() error {
		expired, err := checkCertValidity(sshRunner, cert)
		if err != nil {
			return err
		}
		if !expired {
			return nil
		}
		return &crcerrors.RetriableError{Err: fmt.Errorf("certificate %s still expired", cert)}
	}
}
