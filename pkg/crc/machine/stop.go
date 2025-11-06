package machine

import (
	"os"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/pkg/errors"
)

func (client *client) Stop() (state.State, error) {
	defer func(input, output string) {
		err := cleanKubeconfig(input, output)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			logging.Warnf("Failed to remove crc contexts from kubeconfig: %v", err)
		}
	}(getGlobalKubeConfigPath(), getGlobalKubeConfigPath())
	if running, _ := client.IsRunning(); !running {
		return state.Error, errors.New("Instance is already stopped")
	}
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		return state.Error, errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()
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
