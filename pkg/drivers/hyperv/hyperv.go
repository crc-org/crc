package hyperv

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"time"

	log "github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
)

type Driver struct {
	*drivers.VMDriver
	VirtualSwitch        string
	MacAddress           string
	DisableDynamicMemory bool
}

const (
	defaultMemory               = 8192
	defaultCPU                  = 4
	defaultDisableDynamicMemory = false
)

// NewDriver creates a new Hyper-v driver with default settings.
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		DisableDynamicMemory: defaultDisableDynamicMemory,
		VMDriver: &drivers.VMDriver{
			BaseDriver: &drivers.BaseDriver{
				MachineName: hostName,
				StorePath:   storePath,
			},
			Memory: defaultMemory,
			CPU:    defaultCPU,
		},
	}
}

func (d *Driver) UpdateConfigRaw(rawConfig []byte) error {
	var newDriver Driver

	err := json.Unmarshal(rawConfig, &newDriver)
	if err != nil {
		return err
	}
	if newDriver.Memory != d.Memory {
		log.Debugf("Updating memory from %d MB to %d MB", d.Memory, newDriver.Memory)
		err := cmd("Hyper-V\\Set-VMMemory",
			"-VMName", d.MachineName,
			"-StartupBytes", toMb(newDriver.Memory))
		if err != nil {
			log.Warnf("Failed to update memory to %d MB: %v", newDriver.Memory, err)
			return err
		}
	}

	if newDriver.CPU != d.CPU {
		log.Debugf("Updating CPU count from %d to %d", d.CPU, newDriver.CPU)
		err := cmd("Hyper-V\\Set-VMProcessor",
			d.MachineName,
			"-Count", fmt.Sprintf("%d", newDriver.CPU))
		if err != nil {
			log.Warnf("Failed to set CPU count to %d", newDriver.CPU)
			return err
		}
	}
	if newDriver.DiskCapacity != d.DiskCapacity {
		log.Debugf("Resizing disk from %d bytes to %d bytes", d.DiskCapacity, newDriver.DiskCapacity)
		err := cmd("Hyper-V\\Resize-VHD", "-Path", quote(d.getDiskPath()), "-SizeBytes", fmt.Sprintf("%d", newDriver.DiskCapacity))
		if err != nil {
			log.Warnf("Failed to set disk size to %d", newDriver.DiskCapacity)
			return err
		}
	}
	*d = newDriver
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "hyperv"
}

func (d *Driver) GetState() (state.State, error) {
	stdout, err := cmdOut("(", "Hyper-V\\Get-VM", d.MachineName, ").state")
	if err != nil {
		return state.Error, fmt.Errorf("Failed to find the VM status")
	}

	resp := parseLines(stdout)
	if len(resp) < 1 {
		return state.Error, fmt.Errorf("unexpected Hyper-V state %s", resp)
	}

	switch resp[0] {
	case "Starting", "Running", "Stopping":
		return state.Running, nil
	case "Off":
		return state.Stopped, nil
	default:
		return state.Error, fmt.Errorf("unexpected Hyper-V state %s", resp[0])
	}
}

// PreCreateCheck checks that the machine creation process can be started safely.
func (d *Driver) PreCreateCheck() error {
	// Check that powershell was found
	if _, err := exec.LookPath("powershell.exe"); err != nil {
		return ErrPowerShellNotFound
	}

	// Check that hyperv is installed
	if err := hypervAvailable(); err != nil {
		return err
	}

	// Check that the user is an Administrator
	isAdmin, err := isAdministrator()
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotAdministrator
	}

	if d.VirtualSwitch == "" {
		return nil
	}

	// Check that there is a virtual switch already configured
	if _, err := d.chooseVirtualSwitch(); err != nil {
		return err
	}

	return err
}

func (d *Driver) getDiskPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.%s", d.MachineName, d.ImageFormat))
}

func (d *Driver) Create() error {
	if err := copyFile(d.ImageSourcePath, d.getDiskPath()); err != nil {
		return err
	}

	args := []string{
		"Hyper-V\\New-VM",
		d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(".")),
		"-MemoryStartupBytes", toMb(d.Memory),
	}
	if d.VirtualSwitch != "" {
		virtualSwitch, err := d.chooseVirtualSwitch()
		if err != nil {
			return err
		}
		log.Debugf("Using switch %q", virtualSwitch)
		args = append(args, "-SwitchName", quote(virtualSwitch))
	}

	log.Debugf("Creating VM...")
	if err := cmd(args...); err != nil {
		return err
	}

	if d.VirtualSwitch == "" {
		if err := cmd("Hyper-V\\Remove-VMNetworkAdapter", "-VMName", d.MachineName); err != nil {
			return err
		}
	}

	if d.DisableDynamicMemory {
		if err := cmd("Hyper-V\\Set-VMMemory",
			"-VMName", d.MachineName,
			"-DynamicMemoryEnabled", "$false"); err != nil {
			return err
		}
	}

	if d.CPU > 1 {
		if err := cmd("Hyper-V\\Set-VMProcessor",
			d.MachineName,
			"-Count", fmt.Sprintf("%d", d.CPU)); err != nil {
			return err
		}
	}

	if d.VirtualSwitch != "" && d.MacAddress != "" {
		if err := cmd("Hyper-V\\Set-VMNetworkAdapter",
			"-VMName", d.MachineName,
			"-StaticMacAddress", fmt.Sprintf("\"%s\"", d.MacAddress)); err != nil {
			return err
		}
	}

	if err := cmd("Hyper-V\\Add-VMHardDiskDrive",
		"-VMName", d.MachineName,
		"-Path", quote(d.getDiskPath())); err != nil {
		return err
	}

	return nil
}

func (d *Driver) chooseVirtualSwitch() (string, error) {
	if d.VirtualSwitch == "" {
		return "", errors.New("no virtual switch given")
	}

	stdout, err := cmdOut("[Console]::OutputEncoding = [Text.Encoding]::UTF8; (Hyper-V\\Get-VMSwitch).Name")
	if err != nil {
		return "", err
	}

	switches := parseLines(stdout)

	found := false
	for _, name := range switches {
		if name == d.VirtualSwitch {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("virtual switch %q not found", d.VirtualSwitch)
	}

	return d.VirtualSwitch, nil
}

// waitForIP waits until the host has a valid IP
func (d *Driver) waitForIP() (string, error) {
	if d.VirtualSwitch == "" {
		return "", errors.New("no virtual switch given")
	}

	log.Debugf("Waiting for host to start...")

	for {
		ip, _ := d.GetIP()
		if ip != "" {
			return ip, nil
		}

		time.Sleep(1 * time.Second)
	}
}

// waitStopped waits until the host is stopped
func (d *Driver) waitStopped() error {
	log.Debugf("Waiting for host to stop...")

	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}

		if s != state.Running {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

// Start starts an host
func (d *Driver) Start() error {
	if err := cmd("Hyper-V\\Start-VM", d.MachineName); err != nil {
		return err
	}

	if d.VirtualSwitch == "" {
		return nil
	}

	ip, err := d.waitForIP()
	if err != nil {
		return err
	}

	d.IPAddress = ip

	return nil
}

// Stop stops an host
func (d *Driver) Stop() error {
	if err := cmd("Hyper-V\\Stop-VM", d.MachineName); err != nil {
		return err
	}

	if err := d.waitStopped(); err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

// Remove removes an host
func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}

	return cmd("Hyper-V\\Remove-VM", d.MachineName, "-Force")
}

// Kill force stops an host
func (d *Driver) Kill() error {
	if err := cmd("Hyper-V\\Stop-VM", d.MachineName, "-TurnOff"); err != nil {
		return err
	}

	if err := d.waitStopped(); err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) GetIP() (string, error) {
	if d.VirtualSwitch == "" {
		return "", errors.New("no virtual switch given")
	}

	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	stdout, err := cmdOut("((", "Hyper-V\\Get-VM", d.MachineName, ").networkadapters[0]).ipaddresses[0]")
	if err != nil {
		return "", err
	}

	resp := parseLines(stdout)
	if len(resp) < 1 {
		return "", fmt.Errorf("IP not found")
	}

	return resp[0], nil
}
