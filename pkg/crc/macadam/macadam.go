package macadam

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

type Config struct {
	Runner                crcos.CommandRunner
	MacadamExecutablePath string
	Env                   map[string]string
}

// VMOptions contains all the options for initializing a VM
type VMOptions struct {
	DiskImagePath   string
	DiskSize        uint64
	Memory          uint64
	Name            string
	Username        string
	SSHIdentityPath string
	CPUs            uint64
	CloudInitPath   string
}

// UseMacadam returns the macadam executable configuration
func UseMacadam() Config {
	env := make(map[string]string)
	// Set default environment variables for macadam
	env["CONTAINERS_HELPER_BINARY_DIR"] = constants.CrcBinDir

	return Config{
		Runner:                crcos.NewLocalCommandRunner(),
		MacadamExecutablePath: constants.MacadamPath(),
		Env:                   env,
	}
}

// WithEnv sets environment variables for the macadam command
func (m Config) WithEnv(env map[string]string) Config {
	return Config{
		Runner:                m.Runner,
		MacadamExecutablePath: m.MacadamExecutablePath,
		Env:                   env,
	}
}

// SetEnv sets a single environment variable
func (m Config) SetEnv(key, value string) Config {
	newEnv := make(map[string]string)
	for k, v := range m.Env {
		newEnv[k] = v
	}
	newEnv[key] = value
	return Config{
		Runner:                m.Runner,
		MacadamExecutablePath: m.MacadamExecutablePath,
		Env:                   newEnv,
	}
}

func (m Config) runCommand(isPrivate bool, args ...string) (string, string, error) {
	// Set environment variables and save old values for restoration
	oldEnv := make(map[string]string)
	for key, value := range m.Env {
		if oldValue, exists := os.LookupEnv(key); exists {
			oldEnv[key] = oldValue
		}
		os.Setenv(key, value)
	}

	// Restore environment variables after command execution
	defer func() {
		for key := range m.Env {
			if oldValue, exists := oldEnv[key]; exists {
				os.Setenv(key, oldValue)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	if isPrivate {
		return m.Runner.RunPrivate(m.MacadamExecutablePath, args...)
	}

	return m.Runner.Run(m.MacadamExecutablePath, args...)
}

func (m Config) RunMacadamCommand(args ...string) (string, string, error) {
	return m.runCommand(false, args...)
}

func (m Config) RunMacadamCommandPrivate(args ...string) (string, string, error) {
	return m.runCommand(true, args...)
}

// InitVM initializes a VM using macadam init
func (m Config) InitVM(opts VMOptions) (string, string, error) {
	args := []string{
		"init",
		opts.DiskImagePath,
		"--disk-size", fmt.Sprintf("%d", opts.DiskSize),
		"--memory", fmt.Sprintf("%d", opts.Memory),
		"--name", opts.Name,
		"--username", opts.Username,
		"--ssh-identity-path", opts.SSHIdentityPath,
		"--cpus", fmt.Sprintf("%d", opts.CPUs),
		"--cloud-init", opts.CloudInitPath,
	}
	return m.RunMacadamCommand(args...)
}

// StartVM starts a VM using macadam
func (m Config) StartVM(vmName string) (string, string, error) {
	return m.RunMacadamCommand("start", vmName)
}

// StopVM stops a VM using macadam
func (m Config) StopVM(vmName string) (string, string, error) {
	return m.RunMacadamCommand("stop", vmName)
}

// DeleteVM deletes a VM using macadam
func (m Config) DeleteVM(vmName string) (string, string, error) {
	return m.RunMacadamCommand("rm", "--force", vmName)
}

// ListVMs lists all VMs
func (m Config) ListVMs() (string, string, error) {
	return m.RunMacadamCommand("list")
}

// VMInspectInfo represents the structure returned by macadam inspect
type VMInspectInfo struct {
	ConfigDir struct {
		Path string `json:"Path"`
	} `json:"ConfigDir"`
	Created   string `json:"Created"`
	Name      string `json:"Name"`
	Resources struct {
		CPUs     int   `json:"CPUs"`
		DiskSize int   `json:"DiskSize"`
		Memory   int   `json:"Memory"`
		USBs     []any `json:"USBs"`
	} `json:"Resources"`
	SSHConfig struct {
		IdentityPath   string `json:"IdentityPath"`
		Port           int    `json:"Port"`
		RemoteUsername string `json:"RemoteUsername"`
	} `json:"SSHConfig"`
	State              string `json:"State"`
	UserModeNetworking bool   `json:"UserModeNetworking"`
}

// GetVMStatus gets the status of a VM by inspecting it
func (m Config) GetVMStatus(vmName string) (string, error) {
	stdout, stderr, err := m.RunMacadamCommand("inspect", vmName)
	if err != nil {
		return "", fmt.Errorf("failed to inspect VM: %w (stderr: %s)", err, stderr)
	}

	// Parse the JSON output
	var vms []VMInspectInfo
	if err := json.Unmarshal([]byte(stdout), &vms); err != nil {
		return "", fmt.Errorf("failed to parse inspect output: %w", err)
	}

	if len(vms) == 0 {
		return "", fmt.Errorf("no VM information returned")
	}

	return vms[0].State, nil
}
