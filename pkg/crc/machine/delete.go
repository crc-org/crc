package machine

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/pkg/errors"
)

func (client *client) Delete() error {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
	host, err := libMachineAPIClient.Load(client.name)

	if err != nil {
		return errors.Wrap(err, "Cannot load machine")
	}

	if err := host.Driver.Remove(); err != nil {
		return errors.Wrap(err, "Driver cannot remove machine")
	}

	if err := libMachineAPIClient.Remove(client.name); err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}

	if err := cleanKubeconfig(getGlobalKubeConfigPath(), getGlobalKubeConfigPath()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logging.Warnf("Failed to remove crc contexts from kubeconfig: %v", err)
		}
	}
	return nil
}
