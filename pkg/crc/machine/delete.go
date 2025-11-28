package machine

import (
	"os"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/podman"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/pkg/errors"
)

func (client *client) Delete() error {
	m := getMacadamClient()
	_, _, err := m.DeleteVM(client.name)
	if err != nil {
		return errors.Wrap(err, "Cannot remove machine")
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
