package machine

import (
	"github.com/code-ready/crc/pkg/crc/logging"
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
