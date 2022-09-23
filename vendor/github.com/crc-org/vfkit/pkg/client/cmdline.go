// Package client provides a go API to generate a vfkit commandline.
//
// After creating a `VirtualMachine` object, use its `ToCmdLine()` method to
// get a list of arguments which can be used with the [os/exec] package.
// package client
package client

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Bootloader determines which kernel/initrd/kernel args to use when starting
// the virtual machine. It is mandatory to set a Bootloader or the virtual
// machine won't start.
type Bootloader struct {
	vmlinuzPath   string
	kernelCmdLine string
	initrdPath    string
}

// VirtualMachine is the top-level type. It describes the virtual machine
// configuration (bootloader, devices, ...).
type VirtualMachine struct {
	vcpus       uint
	memoryBytes uint64
	bootloader  *Bootloader
	devices     []VirtioDevice
}

// The VirtioDevice interface is an interface which is implemented by all devices.
type VirtioDevice interface {
	ToCmdLine() ([]string, error)
}

// VirtioVsock configures of a virtio-vsock device allowing 2-way communication
// between the host and the virtual machine type
type VirtioVsock struct {
	// Port is the virtio-vsock port used for this device, see `man vsock` for more
	// details.
	Port uint
	// SocketURL is the path to a unix socket on the host to use for the virtio-vsock communication with the guest.
	SocketURL string
}

// virtioBlk configures a disk device.
type virtioBlk struct {
	imagePath string
}

// virtioRNG configures a random number generator (RNG) device.
type virtioRNG struct {
}

// virtioNet configures the virtual machine networking.
type virtioNet struct {
	nat        bool
	macAddress net.HardwareAddr
}

// virtioSerial configures the virtual machine serial ports.
type virtioSerial struct {
	logFile string
}

// virtioFs configures directory sharing between the guest and the host.
type virtioFs struct {
	sharedDir string
	mountTag  string
}

// NewVirtualMachine creates a new VirtualMachine instance. The virtual machine
// will use vcpus virtual CPUs and it will be allocated memoryBytes bytes of
// RAM. bootloader specifies which kernel/initrd/kernel args it will be using.
func NewVirtualMachine(vcpus uint, memoryBytes uint64, bootloader *Bootloader) *VirtualMachine {
	return &VirtualMachine{
		vcpus:       vcpus,
		memoryBytes: memoryBytes,
		bootloader:  bootloader,
	}
}

// ToCmdLine generates a list of arguments for use with the [os/exec] package.
// These arguments will start a virtual machine with the devices/bootloader/...
// described by vm If the virtual machine configuration described by vm is
// invalid, an error will be returned.
func (vm *VirtualMachine) ToCmdLine() ([]string, error) {
	// TODO: missing binary name/path
	args := []string{}

	if vm.vcpus != 0 {
		args = append(args, "--cpus", strconv.FormatUint(uint64(vm.vcpus), 10))
	}
	if vm.memoryBytes != 0 {
		args = append(args, "--memory", strconv.FormatUint(vm.memoryBytes, 10))
	}

	if vm.bootloader == nil {
		return nil, fmt.Errorf("missing bootloader configuration")
	}
	bootloaderArgs, err := vm.bootloader.ToCmdLine()
	if err != nil {
		return nil, err
	}
	args = append(args, bootloaderArgs...)

	for _, dev := range vm.devices {
		devArgs, err := dev.ToCmdLine()
		if err != nil {
			return nil, err
		}
		args = append(args, devArgs...)
	}

	return args, nil
}

// AddDevice adds a dev to vm. This device can be created with one of the
// VirtioXXXNew methods.
func (vm *VirtualMachine) AddDevice(dev VirtioDevice) error {
	vm.devices = append(vm.devices, dev)

	return nil
}

// NewBootloader creates a new bootloader to start a VM with the file at
// vmlinuzPath as the kernel, kernelCmdLine as the kernel command line, and the
// file at initrdPath as the initrd. The kernel must be uncompressed otherwise
// the VM will fail to boot.
func NewBootloader(vmlinuzPath, kernelCmdLine, initrdPath string) *Bootloader {
	return &Bootloader{
		vmlinuzPath:   vmlinuzPath,
		kernelCmdLine: kernelCmdLine,
		initrdPath:    initrdPath,
	}
}

func (bootloader *Bootloader) ToCmdLine() ([]string, error) {
	args := []string{}
	if bootloader.vmlinuzPath == "" {
		return nil, fmt.Errorf("Missing kernel path")
	}
	args = append(args, "--kernel", bootloader.vmlinuzPath)

	if bootloader.initrdPath == "" {
		return nil, fmt.Errorf("Missing initrd path")
	}
	args = append(args, "--initrd", bootloader.initrdPath)

	if bootloader.kernelCmdLine == "" {
		return nil, fmt.Errorf("Missing kernel command line")
	}
	args = append(args, "--kernel-cmdline", bootloader.kernelCmdLine)

	return args, nil
}

// VirtioVsockNew creates a new virtio-vsock device for 2-way communication
// between the host and the virtual machine. The communication will happen on
// vsock port, and on the host it will use the unix socket at socketURL.
func VirtioVsockNew(port uint, socketURL string) (VirtioDevice, error) {
	return &VirtioVsock{
		Port:      port,
		SocketURL: socketURL,
	}, nil
}

func (dev *VirtioVsock) ToCmdLine() ([]string, error) {
	if dev.Port == 0 || dev.SocketURL == "" {
		return nil, fmt.Errorf("virtio-vsock needs both a port and a socket URL")
	}
	return []string{"--device", fmt.Sprintf("virtio-vsock,port=%d,socketURL=%s", dev.Port, dev.SocketURL)}, nil
}

// VirtioBlkNew creates a new disk to use in the virtual machine. It will use
// the file at imagePath as the disk image. This image must be in raw format.
func VirtioBlkNew(imagePath string) (VirtioDevice, error) {
	return &virtioBlk{
		imagePath: imagePath,
	}, nil
}

func (dev *virtioBlk) ToCmdLine() ([]string, error) {
	if dev.imagePath == "" {
		return nil, fmt.Errorf("virtio-blk needs the path to a disk image")
	}
	return []string{"--device", fmt.Sprintf("virtio-blk,path=%s", dev.imagePath)}, nil
}

// VirtioRNGNew creates a new random number generator device to feed entropy
// into the virtual machine.
func VirtioRNGNew() (VirtioDevice, error) {
	return &virtioRNG{}, nil
}

func (dev *virtioRNG) ToCmdLine() ([]string, error) {
	return []string{"--device", "virtio-rng"}, nil
}

// VirtioNetNew creates a new network device for the virtual machine. It will
// use macAddress as its MAC address.
func VirtioNetNew(macAddress string) (VirtioDevice, error) {
	var hwAddr net.HardwareAddr

	if macAddress != "" {
		var err error
		if hwAddr, err = net.ParseMAC(macAddress); err != nil {
			return nil, err
		}
	}
	return &virtioNet{
		nat:        true,
		macAddress: hwAddr,
	}, nil
}

func (dev *virtioNet) ToCmdLine() ([]string, error) {
	if !dev.nat {
		return nil, fmt.Errorf("virtio-net only support 'nat' networking")
	}
	builder := strings.Builder{}
	builder.WriteString("virtio-net")
	builder.WriteString(",nat")
	if len(dev.macAddress) != 0 {
		builder.WriteString(fmt.Sprintf(",mac=%s", dev.macAddress))
	}

	return []string{"--device", builder.String()}, nil
}

// VirtioSerialNew creates a new serial device for the virtual machine. The
// output the virtual machine sent to the serial port will be written to the
// file at logFilePath.
func VirtioSerialNew(logFilePath string) (VirtioDevice, error) {
	return &virtioSerial{
		logFile: logFilePath,
	}, nil
}

func (dev *virtioSerial) ToCmdLine() ([]string, error) {
	if dev.logFile == "" {
		return nil, fmt.Errorf("virtio-serial needs the path to the log file")
	}
	return []string{"--device", fmt.Sprintf("virtio-serial,logFilePath=%s", dev.logFile)}, nil
}

// VirtioFsNew creates a new virtio-fs device for file sharing. It will share
// the directory at sharedDir with the virtual machine. This directory can be
// mounted in the VM using `mount -t virtiofs mountTag /some/dir`
func VirtioFsNew(sharedDir string, mountTag string) (VirtioDevice, error) {
	return &virtioFs{
		sharedDir: sharedDir,
		mountTag:  mountTag,
	}, nil
}

func (dev *virtioFs) ToCmdLine() ([]string, error) {
	if dev.sharedDir == "" {
		return nil, fmt.Errorf("virtio-fs needs the path to the directory to share")
	}
	if dev.mountTag != "" {
		return []string{"--device", fmt.Sprintf("virtio-fs,sharedDir=%s,mountTag=%s", dev.sharedDir, dev.mountTag)}, nil
	} else {
		return []string{"--device", fmt.Sprintf("virtio-fs,sharedDir=%s", dev.sharedDir)}, nil
	}
}
