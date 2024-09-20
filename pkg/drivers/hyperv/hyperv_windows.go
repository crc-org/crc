package hyperv

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/crc-org/crc/v2/pkg/crc/logging"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
	crcstrings "github.com/crc-org/crc/v2/pkg/strings"
	"github.com/crc-org/machine/libmachine/drivers"
	"github.com/crc-org/machine/libmachine/state"
)

type Driver struct {
	*drivers.VMDriver
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
		if err := d.resizeDisk(int64(newDriver.DiskCapacity)); err != nil {
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
	stdout, stderr, err := powershell.Execute("Hyper-V\\Get-VM", d.MachineName, "|", "Select-Object", "-ExpandProperty", "State")
	if err != nil {
		return state.Error, fmt.Errorf("Failed to find the VM status: %v - %s", err, stderr)
	}

	resp := crcstrings.FirstLine(stdout)
	if resp == "" {
		return state.Error, fmt.Errorf("unexpected Hyper-V state %s", stdout)
	}

	switch resp {
	case "Starting", "Running", "Stopping":
		return state.Running, nil
	case "Off":
		return state.Stopped, nil
	default:
		return state.Error, fmt.Errorf("unexpected Hyper-V state %s", resp)
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

	return nil
}

func (d *Driver) getDiskPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.%s", d.MachineName, d.ImageFormat))
}

func (d *Driver) resizeDisk(newSize int64) error {
	diskPath := d.getDiskPath()
	out, err := cmdOut(fmt.Sprintf("@(Get-VHD -Path %s).Size", quote(diskPath)))
	if err != nil {
		return fmt.Errorf("unable to get current size of crc.vhdx: %w", err)
	}
	currentSize, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		return fmt.Errorf("unable to convert disk size to int: %w", err)
	}
	if newSize == currentSize {
		log.Debugf("%s is already %d bytes", diskPath, newSize)
		return nil
	}
	if newSize < currentSize {
		return fmt.Errorf("current disk image capacity is bigger than the requested size (%d > %d)", currentSize, newSize)
	}

	log.Debugf("Resizing disk from %d bytes to %d bytes", currentSize, newSize)
	return cmd("Hyper-V\\Resize-VHD",
		"-Path",
		quote(diskPath),
		"-SizeBytes",
		fmt.Sprintf("%d", newSize))
}

func (d *Driver) Create() error {
	if err := crcos.CopyFile(d.ImageSourcePath, d.getDiskPath()); err != nil {
		return err
	}

	args := []string{
		"Hyper-V\\New-VM",
		d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(".")),
		"-MemoryStartupBytes", toMb(d.Memory),
	}

	log.Debugf("Creating VM...")
	if err := cmd(args...); err != nil {
		return err
	}

	if err := cmd("Hyper-V\\Remove-VMNetworkAdapter", "-VMName", d.MachineName); err != nil {
		return err
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

	// Disables creating checkpoints and Automatic Start
	// Shuts down the VM when host shuts down
	if err := cmd("Hyper-V\\Set-VM",
		"-VMName", d.MachineName,
		"-AutomaticStartAction", "Nothing",
		"-AutomaticStopAction", "ShutDown",
		"-CheckpointType", "Disabled"); err != nil {
		return err
	}

	if err := cmd("Hyper-V\\Add-VMHardDiskDrive",
		"-VMName", d.MachineName,
		"-Path", quote(d.getDiskPath())); err != nil {
		return err
	}

	return d.resizeDisk(int64(d.DiskCapacity))

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
	if _, _, err := powershell.Execute(`Hyper-V\Get-VM`, d.MachineName, "-ErrorAction", "SilentlyContinue", "-ErrorVariable", "getVmErrors"); err != nil {
		return nil
	}

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

func (d *Driver) GetSharedDirs() ([]drivers.SharedDir, error) {
	for _, dir := range d.SharedDirs {
		if !smbShareExists(dir.Tag) {
			return []drivers.SharedDir{}, nil
		}
	}
	return d.SharedDirs, nil
}
