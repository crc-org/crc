package machine

import (
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/systemd"
	"github.com/pkg/errors"
)

func (client *client) Stop() (state.State, error) {
	if running, _ := client.IsRunning(); !running {
		return state.Error, errors.New("Instance is already stopped")
	}
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		return state.Error, errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()
	if client.GetPreset() == crcPreset.OpenShift {
		if err := stopAllContainers(vm); err != nil {
			logging.Warnf("Failed to stop all OpenShift containers.\nShutting down VM...")
			logging.Debugf("%v", err)
		}
	}
	logging.Info("Stopping the instance, this may take a few minutes...")
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
	// In case usermode networking make sure all the port bind on host should be released
	if client.useVSock() {
		return status, unexposePorts()
	}
	return status, nil
}

// This should be removed after https://bugzilla.redhat.com/show_bug.cgi?id=1965992
// is fixed. We should also ignore the openshift specific errors because stop
// operation shouldn't depend on the openshift side. Without this graceful shutdown
// takes around 6-7 mins.
func stopAllContainers(vm *virtualMachine) error {
	logging.Info("Stopping kubelet and all containers...")
	sshRunner, err := vm.SSHRunner()
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
