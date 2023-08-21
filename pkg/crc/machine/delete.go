package machine

import (
	"os"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/podman"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
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

	// Remove the podman system connection for crc
	if err := podman.RemoveRootlessSystemConnection(); err != nil {
		logging.Debugf("Failed to remove podman rootless system connection: %v", err)
	}
	if err := podman.RemoveRootfulSystemConnection(); err != nil {
		logging.Debugf("Failed to remove podman rootful system connection: %v", err)
	}

	if err := cleanKubeconfig(getGlobalKubeConfigPath(), getGlobalKubeConfigPath()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logging.Warnf("Failed to remove crc contexts from kubeconfig: %v", err)
		}
	}
	return ssh.RemoveCRCHostEntriesFromKnownHosts()
}
