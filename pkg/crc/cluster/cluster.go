package cluster

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/pborman/uuid"
)

// #nosec G101
const vmPullSecretPath = "/var/lib/kubelet/config.json"

func WaitForSSH(sshRunner *ssh.Runner) error {
	checkSSHConnectivity := func() error {
		_, err := sshRunner.Run("exit 0")
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.RetryAfter(300*time.Second, checkSSHConnectivity, time.Second)
}

const (
	kubeletServerCert = "/var/lib/kubelet/pki/kubelet-server-current.pem"
	kubeletClientCert = "/var/lib/kubelet/pki/kubelet-client-current.pem"

	kubeletClientSignerName = "kubernetes.io/kube-apiserver-client-kubelet"

	aggregatorClientCert = "/etc/kubernetes/static-pod-resources/kube-apiserver-certs/configmaps/aggregator-client-ca/ca-bundle.crt"
)

func CheckCertsValidity(sshRunner *ssh.Runner) (bool, bool, error) {
	client, err := checkCertValidity(sshRunner, kubeletClientCert)
	if err != nil {
		return false, false, err
	}
	server, err := checkCertValidity(sshRunner, kubeletServerCert)
	if err != nil {
		return false, false, err
	}
	return client, server, nil
}

func checkCertValidity(sshRunner *ssh.Runner, cert string) (bool, error) {
	output, err := sshRunner.Run(fmt.Sprintf(`date --date="$(sudo openssl x509 -in %s -noout -enddate | cut -d= -f 2)" --iso-8601=seconds`, cert))
	if err != nil {
		return false, err
	}
	expiryDate, err := time.Parse(time.RFC3339, strings.TrimSpace(output))
	if err != nil {
		return false, err
	}
	if time.Now().After(expiryDate) {
		logging.Debugf("Certs have expired, they were valid till: %s", expiryDate.Format(time.RFC822))
		return true, nil
	}
	return false, nil
}

func CheckAggregatorClientCAValidity(sshRunner *ssh.Runner) (bool, error) {
	return checkCertValidity(sshRunner, aggregatorClientCert)
}

// Return size of disk, used space in bytes and the mountpoint
func GetRootPartitionUsage(sshRunner *ssh.Runner) (int64, int64, error) {
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

func EnsurePullSecretPresentInTheCluster(ocConfig oc.Config, pullSec *PullSecret) error {
	if err := WaitForOpenshiftResource(ocConfig, "secret"); err != nil {
		return err
	}

	stdout, _, err := ocConfig.RunOcCommandPrivate("get", "secret", "pull-secret", "-n", "openshift-config", "-o", `jsonpath="{['data']['\.dockerconfigjson']}"`)
	if err != nil {
		return err
	}
	decoded, err := base64.StdEncoding.DecodeString(stdout)
	if err != nil {
		return err
	}
	if err := validation.ImagePullSecret(string(decoded)); err == nil {
		return nil
	}

	logging.Info("Adding user's pull secret to the cluster ...")
	content, err := pullSec.Value()
	if err != nil {
		return err
	}
	base64OfPullSec := base64.StdEncoding.EncodeToString([]byte(content))
	cmdArgs := []string{"patch", "secret", "pull-secret", "-p",
		fmt.Sprintf(`'{"data":{".dockerconfigjson":"%s"}}'`, base64OfPullSec),
		"-n", "openshift-config", "--type", "merge"}

	_, stderr, err := ocConfig.RunOcCommandPrivate(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to add Pull secret %v: %s", err, stderr)
	}
	return nil
}

func EnsureClusterIDIsNotEmpty(ocConfig oc.Config) error {
	if err := WaitForOpenshiftResource(ocConfig, "clusterversion"); err != nil {
		return err
	}

	stdout, _, err := ocConfig.RunOcCommand("get", "clusterversion", "version", "-o", `jsonpath="{['spec']['clusterID']}"`)
	if err != nil {
		return err
	}
	if strings.TrimSpace(stdout) != "" {
		return nil
	}

	logging.Info("Updating cluster ID ...")
	clusterID := uuid.New()
	cmdArgs := []string{"patch", "clusterversion", "version", "-p",
		fmt.Sprintf(`'{"spec":{"clusterID":"%s"}}'`, clusterID), "--type", "merge"}

	_, stderr, err := ocConfig.RunOcCommand(cmdArgs...)
	if err != nil {
		return fmt.Errorf("Failed to update cluster ID %v: %s", err, stderr)
	}

	return nil
}

func AddProxyConfigToCluster(sshRunner *ssh.Runner, ocConfig oc.Config, proxy *network.ProxyConfig) error {
	type trustedCA struct {
		Name string `json:"name"`
	}

	type proxySpecConfig struct {
		HTTPProxy  string    `json:"httpProxy"`
		HTTPSProxy string    `json:"httpsProxy"`
		NoProxy    string    `json:"noProxy"`
		TrustedCA  trustedCA `json:"trustedCA"`
	}

	type patchSpec struct {
		Spec proxySpecConfig `json:"spec"`
	}

	patch := &patchSpec{
		Spec: proxySpecConfig{
			HTTPProxy:  proxy.HTTPProxy,
			HTTPSProxy: proxy.HTTPSProxy,
			NoProxy:    proxy.GetNoProxyString(),
		},
	}

	if err := WaitForOpenshiftResource(ocConfig, "proxy"); err != nil {
		return err
	}

	if proxy.ProxyCACert != "" {
		trustedCAName := "user-ca-bundle"
		logging.Debug("Adding proxy CA cert to cluster")
		if err := addProxyCACertToCluster(sshRunner, ocConfig, proxy, trustedCAName); err != nil {
			return err
		}
		patch.Spec.TrustedCA = trustedCA{Name: trustedCAName}
	}

	patchEncode, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("Failed to encode to json: %v", err)
	}
	logging.Debugf("Patch string %s", string(patchEncode))

	cmdArgs := []string{"patch", "proxy", "cluster", "-p", fmt.Sprintf("'%s'", string(patchEncode)), "-n", "openshift-config", "--type", "merge"}
	if _, stderr, err := ocConfig.RunOcCommand(cmdArgs...); err != nil {
		return fmt.Errorf("Failed to add proxy details %v: %s", err, stderr)
	}
	return nil
}

func addProxyCACertToCluster(sshRunner *ssh.Runner, ocConfig oc.Config, proxy *network.ProxyConfig, trustedCAName string) error {
	proxyConfigMapFileName := fmt.Sprintf("/tmp/%s.json", trustedCAName)
	proxyCABundleTemplate := `{
  "apiVersion": "v1",
  "data": {
    "ca-bundle.crt": "%s"
  },
  "kind": "ConfigMap",
  "metadata": {
    "name": "%s",
    "namespace": "openshift-config"
  }
}
`
	// Replace the carriage return with `\n` char
	p := fmt.Sprintf(proxyCABundleTemplate, strings.ReplaceAll(proxy.ProxyCACert, "\n", `\n`), trustedCAName)
	err := sshRunner.CopyData([]byte(p), proxyConfigMapFileName, 0644)
	if err != nil {
		return err
	}
	cmdArgs := []string{"apply", "-f", proxyConfigMapFileName}
	if _, stderr, err := ocConfig.RunOcCommand(cmdArgs...); err != nil {
		return fmt.Errorf("Failed to add proxy cert details %v: %s", err, stderr)
	}
	return nil
}

// AddProxyToKubeletAndCriO adds the systemd drop-in proxy configuration file to the instance,
// both services (kubelet and crio) need to be restarted after this change.
// Since proxy operator is not able to make changes to in the kubelet/crio side,
// this is the job of machine config operator on the node and for crc this is not
// possible so we do need to put it here.
func AddProxyToKubeletAndCriO(sshRunner *ssh.Runner, proxy *network.ProxyConfig) error {
	proxyTemplate := `[Service]
Environment=HTTP_PROXY=%s
Environment=HTTPS_PROXY=%s
Environment=NO_PROXY=.cluster.local,.svc,10.128.0.0/14,172.30.0.0/16,%s`
	p := fmt.Sprintf(proxyTemplate, proxy.HTTPProxy, proxy.HTTPSProxy, proxy.GetNoProxyString())
	// This will create a systemd drop-in configuration for proxy (both for kubelet and crio services) on the VM.
	err := sshRunner.CopyData([]byte(p), "/etc/systemd/system/crio.service.d/10-default-env.conf", 0644)
	if err != nil {
		return err
	}
	err = sshRunner.CopyData([]byte(p), "/etc/systemd/system/kubelet.service.d/10-default-env.conf", 0644)
	if err != nil {
		return err
	}

	if proxy.ProxyCACert != "" {
		logging.Debug("Adding proxy CA cert to instance")
		return addProxyCACertToInstance(sshRunner, proxy)
	}
	return nil
}

func addProxyCACertToInstance(sshRunner *ssh.Runner, proxy *network.ProxyConfig) error {
	if err := sshRunner.CopyData([]byte(proxy.ProxyCACert), "/etc/pki/ca-trust/source/anchors/openshift-config-user-ca-bundle.crt", 0600); err != nil {
		return err
	}
	if _, err := sshRunner.Run("sudo update-ca-trust"); err != nil {
		return err
	}
	return nil
}

type PullSecret struct {
	value  string
	Getter func() (string, error)
}

func (p *PullSecret) Value() (string, error) {
	if p.value != "" {
		return p.value, nil
	}
	val, err := p.Getter()
	if err == nil {
		p.value = val
	}
	return val, err
}

func EnsurePullSecretPresentOnInstanceDisk(sshRunner *ssh.Runner, pullSecret *PullSecret) error {
	if _, err := sshRunner.Run(fmt.Sprintf("test -e %s", vmPullSecretPath)); err == nil {
		return nil
	}
	logging.Info("Adding user's pull secret to instance disk...")
	content, err := pullSecret.Value()
	if err != nil {
		return err
	}
	return sshRunner.CopyData([]byte(content), vmPullSecretPath, 0600)
}

func WaitforRequestHeaderClientCaFile(ocConfig oc.Config) error {
	if err := WaitForOpenshiftResource(ocConfig, "configmaps"); err != nil {
		return err
	}

	lookupRequestHeaderClientCa := func() error {
		cmdArgs := []string{"get", "configmaps/extension-apiserver-authentication", `-ojsonpath={.data.requestheader-client-ca-file}`,
			"-n", "kube-system"}

		stdout, stderr, err := ocConfig.RunOcCommand(cmdArgs...)
		if err != nil {
			return fmt.Errorf("Failed to get request header client ca file %v: %s", err, stderr)
		}
		if stdout == "" {
			return &errors.RetriableError{Err: fmt.Errorf("missing .data.requestheader-client-ca-file")}
		}
		logging.Debugf("Found .data.requestheader-client-ca-file: %s", stdout)
		return nil
	}
	return errors.RetryAfter(8*time.Minute, lookupRequestHeaderClientCa, 2*time.Second)
}

func DeleteOpenshiftAPIServerPods(ocConfig oc.Config) error {
	if err := WaitForOpenshiftResource(ocConfig, "pod"); err != nil {
		return err
	}

	deleteOpenshiftAPIServerPods := func() error {
		cmdArgs := []string{"delete", "pod", "--all", "-n", "openshift-apiserver"}
		_, _, err := ocConfig.RunOcCommand(cmdArgs...)
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.RetryAfter(60*time.Second, deleteOpenshiftAPIServerPods, time.Second)
}

func CheckProxySettingsForOperator(ocConfig oc.Config, proxy *network.ProxyConfig, deployment, namespace string) (bool, error) {
	if !proxy.IsEnabled() {
		logging.Debugf("No proxy in use")
		return true, nil
	}
	cmdArgs := []string{"set", "env", "deployment", deployment, "--list", "-n", namespace}
	out, _, err := ocConfig.RunOcCommand(cmdArgs...)
	if err != nil {
		return false, err
	}
	if strings.Contains(out, proxy.HTTPSProxy) || strings.Contains(out, proxy.HTTPProxy) {
		return true, nil
	}
	return false, nil
}
