package cloudinit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
)

// UserDataOptions contains all the options needed to generate cloud-init user-data
type UserDataOptions struct {
	PublicKey         string
	PullSecret        string
	KubeAdminPassword string
	DeveloperPassword string
}

const userDataTemplate = `#cloud-config
runcmd:
  - systemctl enable --now kubelet
write_files:
- path: /home/core/.ssh/authorized_keys
  content: '%s'
  owner: core
  permissions: '0600'
- path: /opt/crc/id_rsa.pub
  content: '%s'
  owner: root:root
  permissions: '0644'
- path: /etc/sysconfig/crc-env
  content: |
    CRC_SELF_SUFFICIENT=1
    CRC_NETWORK_MODE_USER=1
  owner: root:root
  permissions: '0644'
- path: /opt/crc/pull-secret
  content: |
    %s
  permissions: '0644'
- path: /opt/crc/pass_kubeadmin
  content: '%s'
  permissions: '0644'
- path: /opt/crc/pass_developer
  content: '%s'
  permissions: '0644'
- path: /opt/crc/ocp-custom-domain.service.done
  permissions: '0644'
`

// compactJSON compacts a JSON string by removing whitespace and newlines
func compactJSON(jsonStr string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(jsonStr)); err != nil {
		return "", fmt.Errorf("failed to compact JSON: %w", err)
	}
	return buf.String(), nil
}

// GenerateUserData generates a cloud-init user-data file and returns the path
func GenerateUserData(machineName string, opts UserDataOptions) (string, error) {
	// Create the machine directory if it doesn't exist
	machineDir := filepath.Dir(getUserDataPath(machineName))
	if err := os.MkdirAll(machineDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create machine directory: %w", err)
	}

	// Compact the pull secret JSON
	compactPullSecret, err := compactJSON(opts.PullSecret)
	if err != nil {
		return "", fmt.Errorf("failed to compact pull secret: %w", err)
	}

	// Generate the cloud-init user-data content
	userData := fmt.Sprintf(userDataTemplate,
		opts.PublicKey,         // /home/core/.ssh/authorized_keys
		opts.PublicKey,         // /opt/crc/id_rsa.pub
		compactPullSecret,      // /opt/crc/pull-secret (compacted)
		opts.KubeAdminPassword, // /opt/crc/pass_kubeadmin
		opts.DeveloperPassword, // /opt/crc/pass_developer
	)

	userDataPath := getUserDataPath(machineName)
	// Write the user-data file
	if err := os.WriteFile(userDataPath, []byte(userData), 0o600); err != nil {
		return "", fmt.Errorf("failed to write user-data file: %w", err)
	}

	return userDataPath, nil
}

// RemoveUserData removes the cloud-init user-data file for a machine
func RemoveUserData(machineName string) error {
	userDataPath := filepath.Join(constants.MachineInstanceDir, machineName, "user-data")
	if err := os.Remove(userDataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove user-data file: %w", err)
	}
	return nil
}

// getUserDataPath returns the path to the user-data file for a machine
func getUserDataPath(machineName string) string {
	return filepath.Join(constants.MachineInstanceDir, machineName, "user-data")
}
