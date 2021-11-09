package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/state"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/pkg/errors"
)

func (client *client) Stop() (state.State, error) {
	vm, err := loadVirtualMachine(client.name)
	if err != nil {
		return state.Error, errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()

	if err := stopAllContainers(vm.Host, client); err != nil {
		logging.Warnf("Failed to stop all OpenShift containers.\nShutting down VM...")
		logging.Debugf("%v", err)
	}
	logging.Info("Stopping the OpenShift cluster, this may take a few minutes...")
	if err := vm.Stop(); err != nil {
		status, stateErr := vm.State()
		if stateErr != nil {
			logging.Debugf("Cannot get VM status after stopping it: %v", stateErr)
		}
		return status, errors.Wrap(err, "Cannot stop machine")
	}
	status, err := vm.State()
	if err != nil {
		return state.Error, errors.Wrap(err, "Cannot get VM status")
	}
	return status, nil
}

// This should be removed after https://bugzilla.redhat.com/show_bug.cgi?id=1965992
// is fixed. We should also ignore the openshift specific errors because stop
// operation shouldn't depend on the openshift side. Without this graceful shutdown
// takes around 6-7 mins.
func stopAllContainers(host *host.Host, client *client) error {
	logging.Info("Stopping kubelet and all containers...")
	instanceIP, err := getIP(host, client.useVSock())
	if err != nil {
		return errors.Wrapf(err, "Error getting the IP")
	}
	sshRunner, err := crcssh.CreateRunner(instanceIP, getSSHPort(client.useVSock()), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath())
	if err != nil {
		return errors.Wrapf(err, "Error creating the ssh client")
	}
	defer sshRunner.Close()

	if err := systemd.NewInstanceSystemdCommander(sshRunner).Stop("kubelet"); err != nil {
		return err
	}
	_, stderr, err := sshRunner.RunPrivileged("stopping all containers", `-- sh -c 'crictl stop $(crictl ps -q)'`)
	if err != nil {
		logging.Errorf("Failed to stop all containers: %v - %s", err, stderr)
		return err
	}
	return nil
}
