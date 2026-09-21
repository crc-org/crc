package machine

import (
	"encoding/json"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/macadam"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/pkg/errors"
)

// getMacadamClient returns a configured macadam client
func getMacadamClient() macadam.Config {
	return macadam.UseMacadam()
}

// vmExists checks if a VM exists by querying macadam
func vmExists(vmName string) (bool, error) {
	m := getMacadamClient()
	stdout, _, err := m.ListVMs()
	if err != nil {
		return false, errors.Wrap(err, "failed to list VMs")
	}

	// Parse the list output to check if VM exists
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, vmName) {
			return true, nil
		}
	}
	return false, nil
}

// getVMState gets the state of a VM using macadam
func getVMState(vmName string) (state.State, error) {
	m := getMacadamClient()
	statusStr, err := m.GetVMStatus(vmName)
	if err != nil {
		// If the command fails, the VM likely doesn't exist or is in error state
		return state.Error, errors.Wrap(err, "failed to get VM status")
	}

	// Parse macadam inspect state output
	// Expected format: "running", "stopped", etc.
	return parseMacadamState(statusStr), nil
}

// parseMacadamState converts macadam status string to CRC state
func parseMacadamState(statusStr string) state.State {
	statusStr = strings.ToLower(strings.TrimSpace(statusStr))

	if strings.Contains(statusStr, "running") {
		return state.Running
	}
	if strings.Contains(statusStr, "stopped") || strings.Contains(statusStr, "shutoff") {
		return state.Stopped
	}
	if strings.Contains(statusStr, "stopping") {
		return state.Stopping
	}
	if strings.Contains(statusStr, "starting") {
		return state.Starting
	}

	return state.Error
}

// getVMIP returns the IP address of the VM
// For now, this returns 127.0.0.1 for vsock mode or attempts to get it from SSH config
func getVMIP(vmName string, useVSock bool) (string, error) {
	if useVSock {
		return "127.0.0.1", nil
	}

	// TODO: Implement IP retrieval for non-vsock mode
	// This might involve reading from macadam config or using SSH
	return "192.168.130.11", nil // Default IP used by CRC
}

// getVMSSHPort returns the SSH port for the VM
func getVMSSHPort(vmName string, useVSock bool) (int, error) {
	if !useVSock {
		return constants.DefaultSSHPort, nil
	}

	// For vsock/user-mode networking, get the dynamic port from macadam inspect
	m := getMacadamClient()
	stdout, stderr, err := m.RunMacadamCommand("inspect", vmName)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to inspect VM (stderr: %s)", stderr)
	}

	// Parse the JSON output
	var vms []macadam.VMInspectInfo
	if err := json.Unmarshal([]byte(stdout), &vms); err != nil {
		return 0, errors.Wrap(err, "failed to parse inspect output")
	}

	if len(vms) == 0 {
		return 0, errors.New("no VM information returned")
	}

	return vms[0].SSHConfig.Port, nil
}
