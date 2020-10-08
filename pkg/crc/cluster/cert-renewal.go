package cluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
)

func waitForPendingCsrs(ocConfig oc.Config, signerName string) error {
	waitForPendingCsr := func() error {
		output, _, err := ocConfig.RunOcCommand("get", "csr")
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		if strings.Contains(output, "Pending") && strings.Contains(output, signerName) {
			return nil
		}
		return &errors.RetriableError{Err: fmt.Errorf("No Pending CSR with signerName %s", signerName)}
	}

	return errors.RetryAfter(8*time.Minute, waitForPendingCsr, time.Second*5)
}

func WaitAndApprovePendingCSRs(ocConfig oc.Config, client, server bool) error {
	/* 2 CSRs to approve, one right after kubelet restart, the other one a few dozen seconds after
	approving the first one
	- First one is requested by system:serviceaccount:openshift-machine-config-operator:node-bootstrapper
	- Second one is requested by system:node:<node_name> */
	if client {
		logging.Info("Kubelet client certificate has expired, renewing it... [will take up to 8 minutes]")
		if err := waitForPendingCsrs(ocConfig, "kubernetes.io/kube-apiserver-client-kubelet"); err != nil {
			logging.Debugf("Error waiting for pending kube-apiserver-client-kubelet CSR: %v", err)
			return err
		}
		if err := ApproveNodeCSR(ocConfig, "kubernetes.io/kube-apiserver-client-kubelet"); err != nil {
			logging.Debugf("Error approving pending kube-apiserver-client-kubelet CSR: %v", err)
			return err
		}
	}
	if server {
		logging.Info("Kubelet serving certificate has expired, waiting for automatic renewal... [will take up to 8 minutes]")
		if err := waitForPendingCsrs(ocConfig, "kubernetes.io/kubelet-serving"); err != nil {
			logging.Debugf("Error waiting for pending kubelet-serving CSR: %v", err)
			return err
		}
		if err := ApproveNodeCSR(ocConfig, "kubernetes.io/kubelet-serving"); err != nil {
			logging.Debugf("Error approving pending kubelet-serving CSR: %v", err)
			return err
		}
	}
	return nil
}
