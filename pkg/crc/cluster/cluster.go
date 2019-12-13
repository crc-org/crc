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

	if err := waitForOpenshiftAPIServer(oc, "secret"); err != nil {
		return err
	}
	_, stderr, err := oc.RunOcCommand(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to add Pull secret %v: %s", err, stderr)
	}
	return nil
}

func UpdateClusterID(oc oc.OcConfig) error {
	clusterID := uuid.New()
	cmdArgs := []string{"patch", "clusterversion", "version", "-p",
		fmt.Sprintf(`{"spec":{"clusterID":"%s"}}`, clusterID), "--type", "merge"}

	if err := waitForOpenshiftAPIServer(oc, "clusterversion"); err != nil {
		return err
	}
	_, stderr, err := oc.RunOcCommand(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to update cluster ID %v: %s", err, stderr)
	}

	return nil
}

func StopAndRemovePodsInVM(sshRunner *ssh.SSHRunner) error {
	// This command make sure we stop the kubelet and clean up the pods
	// We also providing a 2 seconds sleep so that stopped pods get settled and
	// ready for removal. Without this 2 seconds time sometime it happens some of
	// the pods are not completely stopped and when remove happens it will throw
	// an error like below.
	// remove /var/run/containers/storage/overlay-containers/97e5858e610afc9f71d145b1a7bd5ad930e537ccae79969ae256636f7fb7e77c/userdata/shm: device or resource busy
	stopAndRemovePodsCmd := `bash -c 'sudo crictl stopp $(sudo crictl pods -q) && sudo crictl rmp $(sudo crictl pods -q)'`
	stopAndRemovePods := func() error {
		output, err := sshRunner.Run(stopAndRemovePodsCmd)
		logging.Debugf("Output of %s: %s", stopAndRemovePodsCmd, output)
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.RetryAfter(2, stopAndRemovePods, 2*time.Second)
}

func AddProxyConfigToCluster(oc oc.OcConfig, proxy *network.ProxyConfig) error {
	cmdArgs := []string{"patch", "proxy", "cluster", "-p",
		fmt.Sprintf(`{"spec":{"httpProxy":"%s", "httpsProxy":"%s", "noProxy":"%s"}}`, proxy.HttpProxy, proxy.HttpsProxy, proxy.NoProxy),
		"-n", "openshift-config", "--type", "merge"}

	if err := waitForOpenshiftAPIServer(oc, "proxy"); err != nil {
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

	p := fmt.Sprintf(proxyTemplate, proxy.HttpProxy, proxy.HttpsProxy, proxy.NoProxy)
	// This will create a systemd drop-in configuration for proxy (both for kubelet and crio services) on the VM.
	_, err := sshRunner.RunPrivate(fmt.Sprintf("cat <<EOF | sudo tee /etc/systemd/system/crio.service.d/10-default-env.conf\n%s\nEOF", p))
	if err != nil {
		return err
	}
	_, err = sshRunner.RunPrivate(fmt.Sprintf("cat <<EOF | sudo tee /etc/systemd/system/kubelet.service.d/10-default-env.conf\n%s\nEOF", p))
	if err != nil {
		return err
	}
	return nil
}

func addPullSecretToInstanceDisk(sshRunner *ssh.SSHRunner, pullSec string) error {
	_, err := sshRunner.RunPrivate(fmt.Sprintf("cat <<EOF | sudo tee /var/lib/kubelet/config.json\n%s\nEOF", pullSec))
	if err != nil {
		return err
	}
	return nil
}

func waitForOpenshiftAPIServer(oc oc.OcConfig, resource string) error {
	waitForApiServer := func() error {
		stdout, stderr, err := oc.RunOcCommand("get", resource)
		if err != nil {
			logging.Debug(stderr)
			return &errors.RetriableError{Err: err}
		}
		logging.Debug(stdout)
		return nil
	}
	return errors.RetryAfter(80, waitForApiServer, time.Second)
}
