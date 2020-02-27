package machine

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd"
)

var (
	kaoImage = "openshift/cert-recovery"
)

func getRecoveryKubeConfig(output string) (string, error) {
	pattern, err := regexp.Compile(`(?m)export KUBECONFIG=([\w-/.]*)`)
	if err != nil {
		return "", err
	}
	recoveryKubeconfig := []byte{}
	for _, submatches := range pattern.FindAllStringSubmatchIndex(output, -1) {
		recoveryKubeconfig = pattern.ExpandString(recoveryKubeconfig, "$1", output, submatches)
	}
	if len(recoveryKubeconfig) == 0 {
		return "", fmt.Errorf("Could not parse recovery kubeconfig path")
	}

	return string(recoveryKubeconfig), nil
}

type recoveryPod struct {
	sshRunner  *ssh.SSHRunner
	kaoImage   string
	kubeconfig string
}

func (recoveryPod recoveryPod) GetKubeconfigPath() string {
	return recoveryPod.kubeconfig
}

func (recoveryPod recoveryPod) Run(args ...string) (string, string, error) {
	cmd := fmt.Sprintf("sudo oc %s", strings.Join(args, " "))
	logging.Debugf("kube-apiserver-operator: Running %s", cmd)
	stdout, err := recoveryPod.sshRunner.Run(cmd)
	return stdout, "", err
}

func (recoveryPod recoveryPod) RunPrivate(args ...string) (string, string, error) {
	cmd := fmt.Sprintf("sudo oc %s", strings.Join(args, " "))
	stdout, err := recoveryPod.sshRunner.RunPrivate(cmd)
	return stdout, "", err
}

func (recoveryPod *recoveryPod) runPodCommand(cmd string) (string, error) {
	podmanCmd := fmt.Sprintf("sudo podman run -it --rm --network=host -v /etc/kubernetes/:/etc/kubernetes/:Z --entrypoint=/usr/bin/cluster-kube-apiserver-operator '%s'", recoveryPod.kaoImage)
	podmanCmd = podmanCmd + " " + cmd
	return recoveryPod.sshRunner.Run(podmanCmd)
}

func startRecoveryPod(sshRunner *ssh.SSHRunner) (*recoveryPod, error) {
	recoveryPod := recoveryPod{}
	recoveryPod.sshRunner = sshRunner
	recoveryPod.kaoImage = kaoImage
	logging.Debugf("kube-apiserver-operator image: %s", kaoImage)

	output, err := recoveryPod.runPodCommand("recovery-apiserver create")
	if err != nil {
		return nil, err
	}

	logging.Debug("Recovery pod is running")

	recoveryPod.kubeconfig, err = getRecoveryKubeConfig(output)
	if err != nil {
		return nil, err
	}
	logging.Debugf("Recovery KUBECONFIG: [%s]", recoveryPod.kubeconfig)

	return &recoveryPod, nil
}

func waitForApiServer(machineName string) error {
	localOc := oc.UseOCWithConfig(machineName)
	waitForApiServer := func() error {
		_, stderr, err := localOc.RunOcCommand("get", "node")
		if err != nil {
			logging.Debugf("oc invocation failed: %s", stderr)
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.RetryAfter(300, waitForApiServer, time.Second)
}

func waitForRecoveryPod(oc oc.OcConfig) error {
	waitForRecoveryPod := func() error {
		_, _, err := oc.RunOcCommand("get", "namespace", "kube-system")
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}
	return errors.RetryAfter(60, waitForRecoveryPod, time.Second)
}

func waitForPendingCsrs(oc oc.OcConfig) error {
	waitForPendingCsr := func() error {
		output, _, err := oc.RunOcCommand("get", "csr")
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		matched, err := regexp.MatchString("Pending", output)
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		if !matched {
			return &errors.RetriableError{Err: fmt.Errorf("No Pending CSR")}
		}
		return nil
	}

	return errors.RetryAfter(60, waitForPendingCsr, time.Second)
}

func RegenerateCertificates(sshRunner *ssh.SSHRunner, machineName string) error {
	/* FIXME: Need to error out if openshift version is older than 4.1.17 */

	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	startedKubelet, err := sd.Start("kubelet")
	if err != nil {
		logging.Debugf("Error starting kubelet service: %v", err)
		return err
	}
	if startedKubelet {
		defer sd.Stop("kubelet") //nolint:errcheck
	}
	recoveryPod, err := startRecoveryPod(sshRunner)
	if err != nil {
		logging.Debugf("Error starting recovery kube-apiserver-operator: %v", err)
		return err
	}
	defer recoveryPod.runPodCommand("recovery-apiserver destroy") //nolint:errcheck

	oc := oc.NewOcConfig(recoveryPod)
	err = waitForRecoveryPod(oc)
	if err != nil {
		logging.Debugf("Error waiting for recovery apiserver to come up: %v", err)
		return err
	}

	_, err = recoveryPod.runPodCommand("regenerate-certificates")
	if err != nil {
		logging.Debugf("Error running 'regenerate-certificates' command in the kube-apiserver-operator")
		return err
	}

	forceRedeploymentPatch := fmt.Sprintf(`'{"spec": {"forceRedeploymentReason": "recovery-%s"}}'`, time.Now().Format(time.RFC3339))
	for _, resourceType := range []string{"kubeapiserver", "kubecontrollermanager", "kubescheduler"} {
		_, _, err := oc.RunOcCommand(
			"patch", resourceType, "cluster",
			fmt.Sprintf("--patch=%s", forceRedeploymentPatch),
			"--type=merge")
		if err != nil {
			logging.Warnf("Error forcing redeployment of %s: %v", resourceType, err)
			return err
		}
	}

	_, err = sshRunner.Run(fmt.Sprintf("sudo KUBECONFIG=%s /usr/local/bin/recover-kubeconfig.sh >kubeconfig && sudo mv kubeconfig /etc/kubernetes/kubeconfig", recoveryPod.GetKubeconfigPath()))
	if err != nil {
		logging.Debugf("Error invoking recover-kubeconfig.sh: %v", err)
		return err
	}

	output, _, err := oc.RunOcCommand("get", "configmap", "kube-apiserver-to-kubelet-client-ca",
		"-n openshift-kube-apiserver-operator",
		"--template='{{ index .data \"ca-bundle.crt\" }}'")
	if err != nil {
		logging.Debugf("Error reading kubelet client CA: %v", err)
		return err
	}
	/* Copy the ca-bundle data we just read to etc/kubernetes/kubelet-ca.crt */
	err = sshRunner.SetTextContentAsRoot("/etc/kubernetes/kubelet-ca.crt", output, 0644)
	if err != nil {
		logging.Debugf("Error writing kubelet client CA to /etc/kubernetes/kubelet-ca.crt: %v", err)
		return err
	}

	_, _ = sshRunner.Run("sudo touch /run/machine-config-daemon-force")

	_, err = sd.Stop("kubelet")
	if err != nil {
		return err
	}
	_, _ = sshRunner.Run("sudo rm -rf /var/lib/kubelet/pki /var/lib/kubelet/kubeconfig")
	_, _ = sshRunner.Run("sudo crictl stopp $(sudo crictl pods -q)")
	_, _ = sshRunner.Run("sudo crictl rmp $(sudo crictl pods -q)")

	_, err = sd.Start("kubelet")
	if err != nil {
		logging.Debugf("Error restarting kubelet service: %v", err)
		return err
	}

	/* 2 CSRs to approve, one right after kubelet restart, the other one a few dozen seconds after
	approving the first one */
	err = waitForPendingCsrs(oc)
	if err != nil {
		logging.Debugf("Error waiting for first pending CSR: %v", err)
		return err
	}
	err = oc.ApproveNodeCSR()
	if err != nil {
		logging.Debugf("Error approving first pending CSR: %v", err)
		return err
	}

	err = waitForPendingCsrs(oc)
	if err != nil {
		logging.Debugf("Error waiting for second pending CSR: %v", err)
		return err
	}
	err = oc.ApproveNodeCSR()
	if err != nil {
		logging.Debugf("Error approving second pending CSR: %v", err)
		return err
	}

	/* The cluster needs to settle for a few minutes before connections to
	 * api.crc.testing:6443 gets a valid TLS certificate
	 */
	err = waitForApiServer(machineName)
	if err != nil {
		logging.Warnf("API server is not ready after cert recovery process")
	}

	return nil
}
