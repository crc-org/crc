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

	logging.Info("Stopping the instance, this may take a few minutes...")

	m := getMacadamClient()
	_, _, err := m.StopVM(client.name)
	if err != nil {
		return state.Error, errors.Wrap(err, "Cannot stop machine")
	}

	status, err := getVMState(client.name)
	if err != nil {
		return state.Error, errors.Wrap(err, "Cannot get VM status")
	}

	return status, nil
}
