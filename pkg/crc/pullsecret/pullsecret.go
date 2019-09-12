package pullsecret

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/pborman/uuid"
)

var (
	pullSecret = `apiVersion: v1
data:
  .dockerconfigjson: %s
kind: Secret
metadata:
  name: pull-secret
  namespace: openshift-config
type: kubernetes.io/dockerconfigjson`

	// This command timeout in 80 secs if not able to replace the pull secret.
	replacePullSecretCmd = `timeout 80 bash -c 'until oc --config /tmp/kubeconfig replace -f /tmp/pull-secret.yaml 2>/dev/null 1>&2; \
do echo "Waiting for recovery apiserver to come up."; sleep 1; done'`

	// This command is used to update the clusterID using random uuid.
	// https://stackoverflow.com/a/1250279
	updateClusterIdCmd = `timeout 80 bash -c 'until oc --config /tmp/kubeconfig patch clusterversion version -p '"'"'{"spec":{"clusterID":"%s"}}'"'"' --type merge 2>/dev/null 1>&2; \
do echo "Waiting for recovery apiserver to come up."; sleep 1; done'`

	// This command make sure we stop the kubelet and clean up the pods
	// We also providing a 2 sec sleep so that stop pods get settled and
	// ready for remove. Without this 2 sec time sometime it happens some of
	// the pods are not completely stopped and when remove happens it will throw
	// an error like below.
	// remove /var/run/containers/storage/overlay-containers/97e5858e610afc9f71d145b1a7bd5ad930e537ccae79969ae256636f7fb7e77c/userdata/shm: device or resource busy
	stopAndRemovePodsCmd = `bash -c 'sudo crictl stopp $(sudo crictl pods -q) &&\
sudo crictl rmp $(sudo crictl pods -q)'`
)

func AddPullSecretAndClusterID(sshRunner *ssh.SSHRunner, pullSec string, kubeconfigFilePath string) error {
	// Add the kubeconfig File to the Instance.
	if err := addKubeconfigFileToInstance(sshRunner, kubeconfigFilePath); err != nil {
		return err
	}

	// Add the Pull secret kubernetes resource definition to Instance
	if err := addpullSecretSpecToInstance(sshRunner, pullSec); err != nil {
		return err
	}

	// Replace user pull secret and add cluster ID
	if err := setPullSecretAndClusterID(sshRunner); err != nil {
		return err
	}

	// Add user pull secret to the instance
	if err := addpullSecretToInstanceDisk(sshRunner, pullSec); err != nil {
		return err
	}
	return nil
}

func addpullSecretSpecToInstance(sshRunner *ssh.SSHRunner, pullSec string) error {
	base64OfPullSec := base64.StdEncoding.EncodeToString([]byte(pullSec))
	output, err := sshRunner.Run(fmt.Sprintf("cat <<EOF | tee /tmp/pull-secret.yaml\n%s\nEOF", fmt.Sprintf(pullSecret, base64OfPullSec)))
	if err != nil {
		return err
	}
	logging.Debugf("Output is : %s", output)
	return nil
}

func addpullSecretToInstanceDisk(sshRunner *ssh.SSHRunner, pullSec string) error {
	output, err := sshRunner.Run(fmt.Sprintf("cat <<EOF | sudo tee /var/lib/kubelet/config.json\n%s\nEOF", pullSec))
	if err != nil {
		return err
	}
	logging.Debugf("Output is : %s", output)
	return nil
}

func addKubeconfigFileToInstance(sshRunner *ssh.SSHRunner, kubeconfigFilePath string) error {
	kubeconfig, err := ioutil.ReadFile(kubeconfigFilePath)
	if err != nil {
		return err
	}
	_, err = sshRunner.Run(fmt.Sprintf("cat <<EOF | tee /tmp/kubeconfig\n%s\nEOF", string(kubeconfig)))
	if err != nil {
		return err
	}
	return nil
}

func replaceUserPullSecret(sshRunner *ssh.SSHRunner) error {
	output, err := sshRunner.Run(replacePullSecretCmd)
	if err != nil {
		return err
	}
	logging.Debugf("Output of %s: %s", replacePullSecretCmd, output)
	return nil
}

func addClusterID(sshRunner *ssh.SSHRunner) error {
	clusterID := uuid.New()
	updateClusterIdCmd := fmt.Sprintf(updateClusterIdCmd, clusterID)
	output, err := sshRunner.Run(updateClusterIdCmd)
	if err != nil {
		return err
	}
	logging.Debugf("Output of %s: %s", updateClusterIdCmd, output)
	return nil
}

func setPullSecretAndClusterID(sshRunner *ssh.SSHRunner) (rerr error) {
	m := errors.MultiError{}
	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	defer func() {
		// Stop the kubelet service.
		if _, err := sd.Stop("kubelet"); err != nil {
			m.Collect(err)
		}
		stopAndRemovePods := func() error {
			output, err := sshRunner.Run(stopAndRemovePodsCmd)
			logging.Debugf("Output of %s: %s", stopAndRemovePodsCmd, output)
			if err != nil {
				return &errors.RetriableError{Err: err}
			}
			return nil
		}
		if err := errors.RetryAfter(2, stopAndRemovePods, 2*time.Second); err != nil {
			m.Collect(err)
		}
		rerr = m.ToError()
	}()

	// Start the kubelet service
	if _, err := sd.Start("kubelet"); err != nil {
		m.Collect(err)
	}

	// Replace existing pull secret with the user pull secret
	if err := replaceUserPullSecret(sshRunner); err != nil {
		m.Collect(err)
	}

	// add random cluster id
	if err := addClusterID(sshRunner); err != nil {
		m.Collect(err)
	}

	return nil
}
