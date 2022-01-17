package machine

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/pkg/errors"
)

func (client *client) Delete() error {
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil && !errors.Is(err, errInvalidBundleMetadata) {
		return errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()

	if err := vm.Remove(); err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}

	// In case usermode networking make sure all the port bind on host should be released
	if client.useVSock() {
		if err := unexposePorts(); err != nil {
			return err
		}
	}

	if err := cleanKubeconfig(getGlobalKubeConfigPath(), getGlobalKubeConfigPath()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logging.Warnf("Failed to remove crc contexts from kubeconfig: %v", err)
		}
	}
	return nil
}
