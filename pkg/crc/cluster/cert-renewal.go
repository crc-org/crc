package cluster

import (
	"fmt"
	"strings"
	"time"

	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/ssh"
)

func waitForPendingCSRs(ocConfig oc.Config, signerName string) error {
	return crcerrors.RetryAfter(8*time.Minute, func() error {
		output, _, err := ocConfig.RunOcCommand("get", "csr")
		if err != nil {
			return &crcerrors.RetriableError{Err: err}
		}
		if strings.Contains(output, "Pending") && strings.Contains(output, signerName) {
			return nil
		}
		return &crcerrors.RetriableError{Err: fmt.Errorf("No Pending CSR with signerName %s", signerName)}
	}, time.Second*5)
}

func ApproveCSRAndWaitForCertsRenewal(sshRunner *ssh.Runner, ocConfig oc.Config, client, server bool) error {
	// First, kubelet starts and tries to connect to API server. If its certificate is expired, it asks for a new one
	// Admin needs to approve it. The Kubernetes controller manager will then issue the cert, kubelet will fetch it and use it.
	// Kubelet stores the cert in /var/lib/kubelet/pki/kubelet-client-current.pem
	if client {
		logging.Info("Kubelet client certificate has expired, renewing it... [will take up to 8 minutes]")
		if err := waitForPendingCSRs(ocConfig, kubeletClientSignerName); err != nil {
			logging.Debugf("Error waiting for pending kube-apiserver-client-kubelet CSR: %v", err)
			return err
		}
		if err := approveNodeCSR(ocConfig, kubeletClientSignerName); err != nil {
			logging.Debugf("Error approving pending kube-apiserver-client-kubelet CSR: %v", err)
			return err
		}
		if err := crcerrors.RetryAfter(5*time.Minute, waitForCertRenewal(sshRunner, KubeletClientCert), time.Second*5); err != nil {
			logging.Debugf("Error approving pending kube-apiserver-client-kubelet CSR: %v", err)
			return err
		}
	}
	// API server needs to connect to kubelet for some features like logs, port forwards. This communication is backed by a cert
	// store in /var/lib/kubelet/pki/kubelet-server-current.pem
	// After kubelet connected to the API server, if the serving cert is expireed, kubelet asks for a new CSR.
	// This CSR is automatically approved by the cluster-machine-approver. The k8s controller manager issues the cert and kubelet fetches it.
	if server {
		logging.Info("Kubelet serving certificate has expired, waiting for automatic renewal... [will take up to 8 minutes]")
		return crcerrors.RetryAfter(5*time.Minute, waitForCertRenewal(sshRunner, KubeletServerCert), time.Second*5)
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
