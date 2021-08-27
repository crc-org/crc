package cluster

import (
	"fmt"
	"time"

	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/ssh"

	k8scerts "k8s.io/api/certificates/v1beta1"
)

func isPending(csr *k8scerts.CertificateSigningRequest) bool {
	return len(csr.Status.Conditions) == 0 && len(csr.Status.Certificate) == 0
}

func approvePendingCSRs(ocConfig oc.Config, expectedSignerName string) error {
	return crcerrors.RetryAfter(8*time.Minute, func() error {
		csrs, err := getCSRList(ocConfig, expectedSignerName)
		if err != nil {
			return &crcerrors.RetriableError{Err: err}
		}
		csrsApproved := false
		for i := range csrs.Items {
			csr := &csrs.Items[i]
			if !isPending(csr) {
				continue
			}
			logging.Debugf("Approving csr %s (signerName: %s)", csr.ObjectMeta.Name, expectedSignerName)
			_, stderr, err := ocConfig.RunOcCommand("adm", "certificate", "approve", csr.ObjectMeta.Name)
			if err != nil {
				return fmt.Errorf("Not able to approve csr (%v : %s)", err, stderr)
			}
			csrsApproved = true
		}
		if !csrsApproved {
			return &crcerrors.RetriableError{Err: fmt.Errorf("No Pending CSR with signerName %s", expectedSignerName)}
		}
		return nil
	}, time.Second*5)
}

func ApproveCSRAndWaitForCertsRenewal(sshRunner *ssh.Runner, ocConfig oc.Config, client, server bool) error {
	const (
		kubeletClientSignerName = "kubernetes.io/kube-apiserver-client-kubelet"
		authClientSignerName    = "kubernetes.io/kube-apiserver-client"
	)

	// First, kubelet starts and tries to connect to API server. If its certificate is expired, it asks for a new one
	// Admin needs to approve it. The Kubernetes controller manager will then issue the cert, kubelet will fetch it and use it.
	// Kubelet stores the cert in /var/lib/kubelet/pki/kubelet-client-current.pem
	if client {
		logging.Info("Kubelet client certificate has expired, renewing it... [will take up to 8 minutes]")
		if err := approvePendingCSRs(ocConfig, kubeletClientSignerName); err != nil {
			logging.Debugf("Error approving pending kube-apiserver-client-kubelet CSRs: %v", err)
			return err
		}

		// This deleteCSR block only needed for 4.8 version and should be removed when we start shipping 4.9 or
		// if the patch backported to 4.8 z stream.
		// https://github.com/openshift/library-go/pull/1190 and https://github.com/openshift/cluster-authentication-operator/pull/475
		// https://bugzilla.redhat.com/show_bug.cgi?id=1997906
		if err := deleteCSR(ocConfig, authClientSignerName); err != nil {
			logging.Debugf("Error deleting openshift-authenticator csr: %v", err)
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
