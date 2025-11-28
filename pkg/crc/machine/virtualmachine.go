package machine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/pkg/errors"
)

type virtualMachine struct {
	name   string
	bundle *bundle.CrcBundleInfo
	vsock  bool
}

type MissingHostError struct {
	name string
}

func errMissingHost(name string) *MissingHostError {
	return &MissingHostError{name: name}
}

func (err *MissingHostError) Error() string {
	return fmt.Sprintf("no such VM: %s", err.name)
}

var errInvalidBundleMetadata = errors.New("Error loading bundle metadata")

// vmConfig stores metadata about the VM that persists across restarts
type vmConfig struct {
	BundleName string `json:"bundleName"`
}

func loadVirtualMachine(name string, useVSock bool) (*virtualMachine, error) {
	exists, err := vmExists(name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot check if machine exists")
	}
	if !exists {
		return nil, errMissingHost(name)
	}

	crcBundleMetadata, err := getBundleMetadataFromConfig(name)
	if err != nil {
		logging.Debugf("Failed to get bundle metadata: %v", err)
		err = errInvalidBundleMetadata
	}

	return &virtualMachine{
		name:   name,
		bundle: crcBundleMetadata,
		vsock:  useVSock,
	}, err
}

func getBundleMetadataFromConfig(vmName string) (*bundle.CrcBundleInfo, error) {
	// Try to read bundle name from a config file in the machine directory
	configPath := filepath.Join(constants.MachineInstanceDir, vmName, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Fallback: try to get from the most recent bundle
		logging.Debugf("config.json not found, using most recent bundle: %v", err)
		bundles, err := bundle.List()
		if err != nil || len(bundles) == 0 {
			return nil, errors.New("no bundle information available")
		}
		// Use the first/most recent bundle
		return &bundles[0], nil
	}

	// Parse config to get bundle name
	var config vmConfig
	if err := json.Unmarshal(data, &config); err != nil {
		logging.Debugf("Failed to parse config.json, using most recent bundle: %v", err)
		bundles, err := bundle.List()
		if err != nil || len(bundles) == 0 {
			return nil, errors.New("no bundle information available")
		}
		return &bundles[0], nil
	}

	// Get bundle info by name
	return bundle.Get(config.BundleName)
}

// saveBundleMetadataToConfig saves the bundle name to a config file for later retrieval
func saveBundleMetadataToConfig(vmName, bundleName string) error {
	configDir := filepath.Join(constants.MachineInstanceDir, vmName)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	config := vmConfig{
		BundleName: bundleName,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	logging.Debugf("Saved bundle metadata to %s", configPath)
	return nil
}

func (vm *virtualMachine) Close() error {
	// No-op for macadam-based implementation
	return nil
}

func (vm *virtualMachine) Remove() error {
	m := getMacadamClient()
	_, _, err := m.DeleteVM(vm.name)
	if err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}
	return nil
}

func (vm *virtualMachine) State() (state.State, error) {
	return getVMState(vm.name)
}

func (vm *virtualMachine) IP() (string, error) {
	return getVMIP(vm.name, vm.vsock)
}

func (vm *virtualMachine) SSHPort() (int, error) {
	return getVMSSHPort(vm.name, vm.vsock)
}

func (vm *virtualMachine) SSHRunner() (*ssh.Runner, error) {
	ip, err := vm.IP()
	if err != nil {
		return nil, err
	}
	port, err := vm.SSHPort()
	if err != nil {
		return nil, err
	}
	return ssh.CreateRunner(ip, port, constants.GetPrivateKeyPath(), constants.GetECDSAPrivateKeyPath(), vm.bundle.GetSSHKeyPath())
}
