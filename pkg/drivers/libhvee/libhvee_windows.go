package libhvee

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	log "github.com/crc-org/crc/v2/pkg/crc/logging"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/crc-org/machine/libmachine/drivers"
	"github.com/crc-org/machine/libmachine/state"

	"github.com/containers/common/pkg/strongunits"
	"github.com/containers/libhvee/pkg/hypervctl"
)

type Driver struct {
	*drivers.VMDriver
	DynamicMemory bool
}

const (
	defaultMemory        = 8192
	defaultCPU           = 4
	defaultDynamicMemory = false
	systemGeneration     = "Microsoft:Hyper-V:SubType:2"
)

// NewDriver creates a new Hyper-v driver with default settings.
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		VMDriver: &drivers.VMDriver{
			BaseDriver: &drivers.BaseDriver{
				MachineName: hostName,
				StorePath:   storePath,
			},
			Memory: defaultMemory,
			CPU:    defaultCPU,
		},
		DynamicMemory: defaultDynamicMemory,
	}
}

func (d *Driver) UpdateConfigRaw(rawConfig []byte) error {
	var update Driver

	err := json.Unmarshal(rawConfig, &update)
	if err != nil {
		return err
	}

	if update.Memory != d.Memory {
		log.Debugf("Machine: libhvee -> Updating memory from %d MB to %d MB", d.Memory, update.Memory)

		vm, err := d.getMachine()
		if err != nil {
			return err
		}

		err = vm.UpdateProcessorMemSettings(
			func(_ *hypervctl.ProcessorSettings) {},
			func(ms *hypervctl.MemorySettings) {
				ms.VirtualQuantity = uint64(update.Memory)
				ms.Limit = uint64(update.Memory)
			})

		if err != nil {
			return err
		}
	}

	if update.CPU != d.CPU {
		log.Debugf("Machine: libhvee -> Updating CPU count from %d to %d", d.CPU, update.CPU)

		vm, err := d.getMachine()
		if err != nil {
			return err
		}

		err = vm.UpdateProcessorMemSettings(
			func(ps *hypervctl.ProcessorSettings) {
				ps.VirtualQuantity = uint64(update.CPU)
			},
			func(_ *hypervctl.MemorySettings) {})

		if err != nil {
			return err
		}
	}
	if update.DiskCapacity != d.DiskCapacity {
		if err := d.resizeDisk(int64(update.DiskCapacity)); err != nil {
			log.Warnf("Machine: libhvee -> Failed to set disk size to %d", update.DiskCapacity)
			return err
		}
	}
	*d = update
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "libhvee"
}

func (d *Driver) GetState() (state.State, error) {
	log.Debugf("Machine: libhvee -> state")
	vm, err := d.getMachine()
	if err != nil {
		return state.Error, err
	}

	log.Debugf("Machine: libhvee -> state: get")
	vmState := vm.State()
	switch vmState {
	case hypervctl.Enabled:
		log.Debugf("Machine: libhvee -> state: running")
		return state.Running, nil
	case hypervctl.Disabled:
		log.Debugf("Machine: libhvee -> state: stopped")
		return state.Stopped, nil
	}

	log.Debugf("Machine: libhvee -> state: unknown")
	return state.Error, fmt.Errorf("unknown state")
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

func (d *Driver) Create() error {
	log.Debugf("Machine: libhvee -> creating: system settings")
	systemSettings, err := hypervctl.NewSystemSettingsBuilder().
		PrepareSystemSettings(d.MachineName, nil).
		PrepareMemorySettings(
			func(ms *hypervctl.MemorySettings) {
				ms.DynamicMemoryEnabled = d.DynamicMemory
				ms.VirtualQuantity = uint64(d.Memory)
				ms.Reservation = 1024
				ms.Limit = uint64(d.Memory)
			}).
		PrepareProcessorSettings(
			func(ps *hypervctl.ProcessorSettings) {
				ps.VirtualQuantity = uint64(d.CPU)
			}).
		Build()

	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> creating: copy disk image")
	diskPath := d.getDiskPath()
	if err := crcos.CopyFile(d.ImageSourcePath, diskPath); err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> creating: hardware setup")
	err = hypervctl.NewDriveSettingsBuilder(systemSettings).
		AddScsiController().
		AddSyntheticDiskDrive(0).
		DefineVirtualHardDisk(diskPath,
			func(_ *hypervctl.VirtualHardDiskStorageSettings) {}).
		Finish().
		Finish().
		Finish().
		Complete()

	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> creating: done")

	return d.resizeDisk(int64(d.DiskCapacity))

}

func (d *Driver) getMachine() (*hypervctl.VirtualMachine, error) {
	vmm := hypervctl.NewVirtualMachineManager()

	log.Debugf("Machine: libhvee -> get machine")
	return vmm.GetMachine(d.MachineName)
}

// Start starts an host
func (d *Driver) Start() error {
	vm, err := d.getMachine()
	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> start")
	return vm.Start()
}

// Stop stops an host
func (d *Driver) Stop() error {
	vm, err := d.getMachine()
	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> stop")
	return vm.Stop()
}

// Remove removes an host
func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> remove running")
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}

	vm, err := d.getMachine()
	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> remove vm")
	err = vm.Remove("")
	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> remove disk file")
	return os.Remove(d.getDiskPath())
}

// Kill force stops an host
func (d *Driver) Kill() error {
	vm, err := d.getMachine()
	if err != nil {
		return err
	}

	log.Debugf("Machine: libhvee -> stop with force")
	return vm.StopWithForce()
}

func (d *Driver) GetIP() (string, error) {
	return "", drivers.ErrNotSupported
}

func (d *Driver) GetSharedDirs() ([]drivers.SharedDir, error) {
	for _, dir := range d.SharedDirs {
		if !smbShareExists(dir.Tag) {
			return []drivers.SharedDir{}, nil
		}
	}
	return d.SharedDirs, nil
}

func (d *Driver) getDiskPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.%s", d.MachineName, d.ImageFormat))
}

func (d *Driver) resizeDisk(newSizeBytes int64) error {

	newSize := strongunits.B(newSizeBytes)

	diskPath := d.getDiskPath()
	currentSize, err := hypervctl.GetDiskSize(diskPath)
	if err != nil {
		return fmt.Errorf("unable to get current size of crc.vhdx: %w", err)
	}
	if newSize == currentSize.ToBytes() {
		log.Debugf("%s is already %d bytes", diskPath, newSize)
		return nil
	}
	if newSize < currentSize.ToBytes() {
		return fmt.Errorf("current disk image capacity is bigger than the requested size (%d > %d)", currentSize.ToBytes(), newSize)
	}

	log.Debugf("Resizing disk from %d bytes to %d bytes", currentSize.ToBytes(), newSize.ToBytes())
	return hypervctl.ResizeDisk(diskPath, strongunits.ToGiB(newSize))
}
