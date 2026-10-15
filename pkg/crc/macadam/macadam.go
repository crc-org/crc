package macadam

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
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
	SharedDirs      []string
}

// ConfigVMOptions contains the updatable VM settings for macadam set
type ConfigVMOptions struct {
	Name     string
	CPUs     uint64
	Memory   uint64 // Memory in MiB
	DiskSize uint64 // Disk size in GiB
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

func (m Config) withGlobalFlags(args []string) []string {
	// macadam global options come before the subcommand:
	// macadam --log-level info [--provider hyperv] <command> ...
	global := []string{"--log-level", logging.DefaultLogLevel().String()}
	if runtime.GOOS == "windows" {
		global = append(global, "--provider", "hyperv")
	}
	return append(global, args...)
}

func (m Config) RunMacadamCommand(args ...string) (string, string, error) {
	return m.runCommand(false, m.withGlobalFlags(args)...)
}

func (m Config) RunMacadamCommandPrivate(args ...string) (string, string, error) {
	return m.runCommand(true, m.withGlobalFlags(args)...)
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
	if runtime.GOOS == "windows" {
		args = append(args, "--user-mode-networking")
	}
	// TODO: shared dir not implemented yet
	// https://github.com/crc-org/macadam/issues/407
	/*if len(opts.SharedDirs) > 0 {
		for _, sharedDir := range opts.SharedDirs {
			args = append(args, "--volume", sharedDir)
		}
	}*/
	return m.RunMacadamCommand(args...)
}

// ConfigVM updates VM settings using macadam set
func (m Config) ConfigVM(opts ConfigVMOptions) (string, string, error) {
	vmInfo, err := m.InspectVM(opts.Name)
	if err != nil {
		return "", "", err
	}

	args := []string{
		"set",
		"--cpus", fmt.Sprintf("%d", opts.CPUs),
		"--memory", fmt.Sprintf("%d", opts.Memory),
	}
	if opts.DiskSize > vmInfo.Resources.DiskSize {
		args = append(args, "--disk-size", fmt.Sprintf("%d", opts.DiskSize))
	}
	args = append(args, opts.Name)
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
		CPUs     uint64 `json:"CPUs"`
		DiskSize uint64 `json:"DiskSize"`
		Memory   uint64 `json:"Memory"`
		USBs     []any  `json:"USBs"`
	} `json:"Resources"`
	SSHConfig struct {
		IdentityPath   string `json:"IdentityPath"`
		Port           int    `json:"Port"`
		RemoteUsername string `json:"RemoteUsername"`
	} `json:"SSHConfig"`
	State              string `json:"State"`
	UserModeNetworking bool   `json:"UserModeNetworking"`
	Services           struct {
		GVProxy struct {
			ServiceSocket struct {
				Path string `json:"Path"`
			} `json:"ServiceSocket"`
			Logs struct {
				Path string `json:"Path"`
			} `json:"Logs"`
		} `json:"GVProxy"`
	} `json:"Services"`
}

// InspectVM inspects a VM and returns its information
func (m Config) InspectVM(vmName string) (*VMInspectInfo, error) {
	stdout, stderr, err := m.RunMacadamCommand("inspect", vmName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect VM: %w (stderr: %s)", err, stderr)
	}

	var vms []VMInspectInfo
	if err := json.Unmarshal([]byte(stdout), &vms); err != nil {
		return nil, fmt.Errorf("failed to parse inspect output: %w", err)
	}

	if len(vms) == 0 {
		return nil, fmt.Errorf("no VM information returned")
	}

	return &vms[0], nil
}

// GetVMStatus gets the status of a VM by inspecting it
func (m Config) GetVMStatus(vmName string) (string, error) {
	vmInfo, err := m.InspectVM(vmName)
	if err != nil {
		return "", err
	}
	return vmInfo.State, nil
}

// GetGVProxySocketPath returns the gvproxy socket path for a VM
func (m Config) GetGVProxySocketPath(vmName string) (string, error) {
	vmInfo, err := m.InspectVM(vmName)
	if err != nil {
		return "", err
	}
	socketPath := vmInfo.Services.GVProxy.ServiceSocket.Path
	if socketPath == "" {
		return "", fmt.Errorf("gvproxy socket path not found in VM info")
	}
	return socketPath, nil
}
