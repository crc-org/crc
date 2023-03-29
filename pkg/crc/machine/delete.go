package machine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/pkg/crc/constants"
	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/crc-org/crc/pkg/crc/podman"
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
	return removeCRCHostEntriesFromKnownHosts()
}

func removeCRCHostEntriesFromKnownHosts() error {
	knownHostsPath := filepath.Join(constants.GetHomeDir(), ".ssh", "known_hosts")
	if _, err := os.Stat(knownHostsPath); err != nil {
		return nil
	}
	f, err := os.Open(knownHostsPath)
	if err != nil {
		return fmt.Errorf("Unable to open user's 'known_hosts' file: %w", err)
	}
	defer f.Close()

	tempHostsFile, err := os.CreateTemp(filepath.Join(constants.GetHomeDir(), ".ssh"), "crc")
	if err != nil {
		return fmt.Errorf("Unable to create temp file: %w", err)
	}
	defer func() {
		tempHostsFile.Close()
		os.Remove(tempHostsFile.Name())
	}()

	if err := tempHostsFile.Chmod(0600); err != nil {
		return fmt.Errorf("Error trying to change permissions for temp file: %w", err)
	}

	var foundCRCEntries bool
	scanner := bufio.NewScanner(f)
	writer := bufio.NewWriter(tempHostsFile)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "[127.0.0.1]:2222") || strings.Contains(scanner.Text(), "192.168.130.11") {
			foundCRCEntries = true
			continue
		}
		if _, err := writer.WriteString(fmt.Sprintf("%s\n", scanner.Text())); err != nil {
			return fmt.Errorf("Error while writing hostsfile content to temp file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Error while flushing buffered content to temp file: %w", err)
	}

	if foundCRCEntries {
		if err := f.Close(); err != nil {
			return fmt.Errorf("Error closing known_hosts file: %w", err)
		}
		if err := tempHostsFile.Close(); err != nil {
			return fmt.Errorf("Error closing temp file: %w", err)
		}
		return os.Rename(tempHostsFile.Name(), knownHostsPath)
	}
	return nil
}
