package hyperv

import (
	"fmt"
	"net"
	"time"

	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/log"
	"github.com/code-ready/machine/libmachine/mcnflag"
	"github.com/code-ready/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	CrcDiskCopier        CRCDiskCopier
	VirtualSwitch        string
	DiskPath             string
	DiskPathUrl          string
	Memory               int
	CPU                  int
	MacAddress           string
	DisableDynamicMemory bool
	SSHKeyPath           string
}

const (
	defaultMemory               = 8192
	defaultCPU                  = 4
	defaultDisableDynamicMemory = false
)

// NewDriver creates a new Hyper-v driver with default settings.
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		CrcDiskCopier:        NewCRCDiskCopier(),
		Memory:               defaultMemory,
		CPU:                  defaultCPU,
		DisableDynamicMemory: defaultDisableDynamicMemory,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "hyperv-bundlepath-url",
			Usage:  "URL of the crc bundlepath. Defaults to the latest available version.",
			EnvVar: "HYPERV_BUNDLEPATH_URL",
		},
		mcnflag.StringFlag{
			Name:   "hyperv-virtual-switch",
			Usage:  "Virtual switch name. Defaults to first found.",
			EnvVar: "HYPERV_VIRTUAL_SWITCH",
		},
		mcnflag.IntFlag{
			Name:   "hyperv-memory",
			Usage:  "Memory size for host in MB.",
			Value:  defaultMemory,
			EnvVar: "HYPERV_MEMORY",
		},
		mcnflag.IntFlag{
			Name:   "hyperv-cpu-count",
			Usage:  "number of CPUs for the machine",
			Value:  defaultCPU,
			EnvVar: "HYPERV_CPU_COUNT",
		},
		mcnflag.StringFlag{
			Name:   "hyperv-static-macaddress",
			Usage:  "Hyper-V network adapter's static MAC address.",
			EnvVar: "HYPERV_STATIC_MACADDRESS",
		},
		mcnflag.BoolFlag{
			Name:   "hyperv-disable-dynamic-memory",
			Usage:  "Disable dynamic memory management setting",
			EnvVar: "HYPERV_DISABLE_DYNAMIC_MEMORY",
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.VirtualSwitch = flags.String("hyperv-virtual-switch")
	d.Memory = flags.Int("hyperv-memory")
	d.CPU = flags.Int("hyperv-cpu-count")
	d.MacAddress = flags.String("hyperv-static-macaddress")
	d.SSHUser = drivers.DefaultSSHUser
	d.DisableDynamicMemory = flags.Bool("hyperv-disable-dynamic-memory")

	return nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return d.SSHKeyPath
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "hyperv"
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	stdout, err := cmdOut("(", "Hyper-V\\Get-VM", d.MachineName, ").state")
	if err != nil {
		return state.None, fmt.Errorf("Failed to find the VM status")
	}

	resp := parseLines(stdout)
	if len(resp) < 1 {
		return state.None, nil
	}

	switch resp[0] {
	case "Running":
		return state.Running, nil
	case "Off":
		return state.Stopped, nil
	default:
		return state.None, nil
	}
}

// PreCreateCheck checks that the machine creation process can be started safely.
func (d *Driver) PreCreateCheck() error {
	// Check that powershell was found
	if powershell == "" {
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

	// Check that there is a virtual switch already configured
	if _, err := d.chooseVirtualSwitch(); err != nil {
		return err
	}

	return err
}

func (d *Driver) Create() error {
	if err := d.CrcDiskCopier.CopyDiskToMachineDir(d.StorePath, d.MachineName, d.DiskPathUrl); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	virtualSwitch, err := d.chooseVirtualSwitch()
	if err != nil {
		return err
	}

	log.Infof("Using switch %q", virtualSwitch)

	if err := cmd("Hyper-V\\New-VM",
		d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(".")),
		"-SwitchName", quote(virtualSwitch),
		"-MemoryStartupBytes", toMb(d.Memory)); err != nil {
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

	if d.MacAddress != "" {
		if err := cmd("Hyper-V\\Set-VMNetworkAdapter",
			"-VMName", d.MachineName,
			"-StaticMacAddress", fmt.Sprintf("\"%s\"", d.MacAddress)); err != nil {
			return err
		}
	}

	if err := cmd("Hyper-V\\Add-VMHardDiskDrive",
		"-VMName", d.MachineName,
		"-Path", quote(d.DiskPath)); err != nil {
		return err
	}

	log.Infof("Starting VM...")
	return d.Start()
}

func (d *Driver) chooseVirtualSwitch() (string, error) {
	if d.VirtualSwitch == "" {
		// Default to the first external switch and in the process avoid DockerNAT
		stdout, err := cmdOut("[Console]::OutputEncoding = [Text.Encoding]::UTF8; (Hyper-V\\Get-VMSwitch -SwitchType External).Name")
		if err != nil {
			return "", err
		}

		switches := parseLines(stdout)

		if len(switches) < 1 {
			return "", fmt.Errorf("no external virtual switch found. A valid virtual switch must be available for this command to run")
		}

		return switches[0], nil
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
	log.Infof("Waiting for host to start...")

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
	log.Infof("Waiting for host to stop...")

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

// Restart stops and starts an host
func (d *Driver) Restart() error {
	err := d.Stop()
	if err != nil {
		return err
	}

	return d.Start()
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
