package cluster

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/pborman/uuid"
)

func WaitForSsh(sshRunner *ssh.SSHRunner) error {
	checkSshConnectivity := func() error {
		_, err := sshRunner.Run("exit 0")
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.RetryAfter(60, checkSshConnectivity, time.Second)
}

type CertExpiryState int

const (
	Unknown CertExpiryState = iota
	CertNotExpired
	CertExpired
)

// CheckCertsValidity checks if the cluster certs have expired or going to expire in next 7 days
func CheckCertsValidity(sshRunner *ssh.SSHRunner) (CertExpiryState, error) {
	certExpiryDate, err := getcertExpiryDateFromVM(sshRunner)
	if err != nil {
		return Unknown, err
	}
	if time.Now().After(certExpiryDate) {
		return CertExpired, fmt.Errorf("Certs have expired, they were valid till: %s", certExpiryDate.Format(time.RFC822))
	}

	return CertNotExpired, nil
}

func getcertExpiryDateFromVM(sshRunner *ssh.SSHRunner) (time.Time, error) {
	certExpiryDate := time.Time{}
	certExpiryDateCmd := `date --date="$(sudo openssl x509 -in /var/lib/kubelet/pki/kubelet-client-current.pem -noout -enddate | cut -d= -f 2)" --iso-8601=seconds`
	output, err := sshRunner.Run(certExpiryDateCmd)
	if err != nil {
		return certExpiryDate, err
	}
	certExpiryDate, err = time.Parse(time.RFC3339, strings.TrimSpace(output))
	if err != nil {
		return certExpiryDate, err
	}
	return certExpiryDate, nil
}

// Return size of disk, used space in bytes and the mountpoint
func GetRootPartitionUsage(sshRunner *ssh.SSHRunner) (int64, int64, error) {
	cmd := "df -B1 --output=size,used,target /sysroot | tail -1"

	out, err := sshRunner.Run(cmd)

	if err != nil {
		return 0, 0, err
	}
	diskDetails := strings.Split(strings.TrimSpace(out), " ")
	diskSize, err := strconv.ParseInt(diskDetails[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	diskUsage, err := strconv.ParseInt(diskDetails[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return diskSize, diskUsage, nil
}

func AddPullSecret(sshRunner *ssh.SSHRunner, oc oc.OcConfig, pullSec string) error {
	if err := addPullSecretToInstanceDisk(sshRunner, pullSec); err != nil {
		return err
	}

	base64OfPullSec := base64.StdEncoding.EncodeToString([]byte(pullSec))
	cmdArgs := []string{"patch", "secret", "pull-secret", "-p",
		fmt.Sprintf(`{"data":{".dockerconfigjson":"%s"}}`, base64OfPullSec),
		"-n", "openshift-config", "--type", "merge"}

	if err := oc.WaitForOpenshiftResource("secret"); err != nil {
		return err
	}
	_, stderr, err := oc.RunOcCommandPrivate(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to add Pull secret %v: %s", err, stderr)
	}
	return nil
}

func UpdateClusterID(oc oc.OcConfig) error {
	clusterID := uuid.New()
	cmdArgs := []string{"patch", "clusterversion", "version", "-p",
		fmt.Sprintf(`{"spec":{"clusterID":"%s"}}`, clusterID), "--type", "merge"}

	if err := oc.WaitForOpenshiftResource("clusterversion"); err != nil {
		return err
	}
	_, stderr, err := oc.RunOcCommand(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to update cluster ID %v: %s", err, stderr)
	}

	return nil
}

func AddProxyConfigToCluster(oc oc.OcConfig, proxy *network.ProxyConfig) error {
	cmdArgs := []string{"patch", "proxy", "cluster", "-p",
		fmt.Sprintf(`{"spec":{"httpProxy":"%s", "httpsProxy":"%s", "noProxy":"%s"}}`, proxy.HttpProxy, proxy.HttpsProxy, proxy.GetNoProxyString()),
		"-n", "openshift-config", "--type", "merge"}

	if err := oc.WaitForOpenshiftResource("proxy"); err != nil {
		return err
	}
	if _, stderr, err := oc.RunOcCommand(cmdArgs...); err != nil {
		return fmt.Errorf("Failed to add proxy details %v: %s", err, stderr)
	}
	return nil
}

// AddProxyToKubeletAndCriO adds the systemd drop-in proxy configuration file to the instance,
// both services (kubelet and crio) need to be restarted after this change.
// Since proxy operator is not able to make changes to in the kubelet/crio side,
// this is the job of machine config operator on the node and for crc this is not
// possible so we do need to put it here.
func AddProxyToKubeletAndCriO(sshRunner *ssh.SSHRunner, proxy *network.ProxyConfig) error {
	proxyTemplate := `[Service]
Environment=HTTP_PROXY=%s
Environment=HTTPS_PROXY=%s
Environment=NO_PROXY=.cluster.local,.svc,10.128.0.0/14,172.30.0.0/16,%s`
	p := fmt.Sprintf(proxyTemplate, proxy.HttpProxy, proxy.HttpsProxy, proxy.GetNoProxyString())
	// This will create a systemd drop-in configuration for proxy (both for kubelet and crio services) on the VM.
	err := sshRunner.SetTextContentAsRoot("/etc/systemd/system/crio.service.d/10-default-env.conf", p, 0644)
	if err != nil {
		return err
	}
	err = sshRunner.SetTextContentAsRoot("/etc/systemd/system/kubelet.service.d/10-default-env.conf", p, 0644)
	if err != nil {
		return err
	}
	return nil
}

func addPullSecretToInstanceDisk(sshRunner *ssh.SSHRunner, pullSec string) error {
	err := sshRunner.SetTextContentAsRoot("/var/lib/kubelet/config.json", pullSec, 0600)
	if err != nil {
		return err
	}

	return nil
}

func WaitforRequestHeaderClientCaFile(oc oc.OcConfig) error {
	if err := oc.WaitForOpenshiftResource("configmaps"); err != nil {
		return err
	}

	lookupRequestHeaderClientCa := func() error {
		cmdArgs := []string{"get", "configmaps/extension-apiserver-authentication", `-ojsonpath={.data.requestheader-client-ca-file}`,
			"-n", "kube-system"}

		stdout, stderr, err := oc.RunOcCommand(cmdArgs...)
		if err != nil {
			return fmt.Errorf("Failed to get request header client ca file %v: %s", err, stderr)
		}
		if stdout == "" {
			return &errors.RetriableError{Err: fmt.Errorf("missing .data.requestheader-client-ca-file")}
		}
		logging.Debugf("Found .data.requestheader-client-ca-file: %s", stdout)
		return nil
	}
	return errors.RetryAfter(90, lookupRequestHeaderClientCa, 2*time.Second)
}

func DeleteOpenshiftApiServerPods(oc oc.OcConfig) error {
	if err := oc.WaitForOpenshiftResource("pod"); err != nil {
		return err
	}

	deleteOpenshiftApiserverPods := func() error {
		cmdArgs := []string{"delete", "pod", "--all", "-n", "openshift-apiserver"}
		_, _, err := oc.RunOcCommand(cmdArgs...)
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.RetryAfter(60, deleteOpenshiftApiserverPods, time.Second)
}

func CheckProxySettingsForOperator(oc oc.OcConfig, proxy *network.ProxyConfig, deployment, namespace string) (bool, error) {
	if !proxy.IsEnabled() {
		logging.Debugf("No proxy in use")
		return true, nil
	}
	cmdArgs := []string{"set", "env", "deployment", deployment, "--list", "-n", namespace}
	out, _, err := oc.RunOcCommand(cmdArgs...)
	if err != nil {
		return false, err
	}
	if strings.Contains(out, proxy.HttpsProxy) || strings.Contains(out, proxy.HttpProxy) {
		return true, nil
	}
	return false, nil
}
