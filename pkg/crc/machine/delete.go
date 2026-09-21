package machine

import (
	"os"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/pkg/errors"
)

func (client *client) Delete() error {
	m := getMacadamClient()
	_, _, err := m.DeleteVM(client.name)
	if err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}

	if err := cleanKubeconfig(getGlobalKubeConfigPath(), getGlobalKubeConfigPath()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logging.Warnf("Failed to remove crc contexts from kubeconfig: %v", err)
		}
	}
	return ssh.RemoveCRCHostEntriesFromKnownHosts()
}
