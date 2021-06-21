package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/oc"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/pkg/errors"
)

func (client *client) Stop() (state.State, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
	host, err := libMachineAPIClient.Load(client.name)

	if err != nil {
		return state.Error, errors.Wrap(err, "Cannot load machine")
	}
	if err := removeMCOPods(host, client); err != nil {
		return state.Error, err
	}
	logging.Info("Stopping the OpenShift cluster, this may take a few minutes...")
	if err := host.Stop(); err != nil {
		status, stateErr := host.Driver.GetState()
		if stateErr != nil {
			logging.Debugf("Cannot get VM status after stopping it: %v", stateErr)
		}
		return status, errors.Wrap(err, "Cannot stop machine")
	}
	return host.Driver.GetState()
}

// This should be removed after https://bugzilla.redhat.com/show_bug.cgi?id=1965992
// is fixed. We should also ignore the openshift specific errors because stop
// operation shouldn't depend on the openshift side. Without this graceful shutdown
// takes around 6-7 mins.
func removeMCOPods(host *host.Host, client *client) error {
	logging.Info("Deleting the pods from openshift-machine-config-operator namespace")
	instanceIP, err := getIP(host, client.useVSock())
	if err != nil {
		return errors.Wrapf(err, "Error getting the IP")
	}
	sshRunner, err := crcssh.CreateRunner(instanceIP, getSSHPort(client.useVSock()), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath())
	if err != nil {
		return errors.Wrapf(err, "Error creating the ssh client")
	}
	defer sshRunner.Close()

	ocConfig := oc.UseOCWithSSH(sshRunner)
	_, stderr, err := ocConfig.RunOcCommand("delete", "pods", "--all", "-n openshift-machine-config-operator", "--grace-period=0")
	if err != nil {
		logging.Debugf("Error deleting the pods from openshift-machine-config-operator namespace:%v \n StdErr: %s", err, stderr)
	}
	return nil
}
