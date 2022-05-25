package client

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Bootloader struct {
	vmlinuzPath   string
	kernelCmdLine string
	initrdPath    string
}

type VirtualMachine struct {
	vcpus       uint
	memoryBytes uint64
	bootloader  *Bootloader
	devices     []VirtioDevice
}

type VirtioDevice interface {
	ToCmdLine() ([]string, error)
}

type VirtioVsock struct {
	Port      uint
	SocketURL string
}

type virtioBlk struct {
	imagePath string
}

type virtioRNG struct {
}

type virtioNet struct {
	nat        bool
	macAddress net.HardwareAddr
}

type virtioSerial struct {
	logFile string
}

type virtioFs struct {
	sharedDir string
	mountTag  string
}

func NewVirtualMachine(vcpus uint, memoryBytes uint64, bootloader *Bootloader) *VirtualMachine {
	return &VirtualMachine{
		vcpus:       vcpus,
		memoryBytes: memoryBytes,
		bootloader:  bootloader,
	}
}

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

func (vm *VirtualMachine) AddDevice(dev VirtioDevice) error {
	vm.devices = append(vm.devices, dev)

	return nil
}

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

func VirtioRNGNew() (VirtioDevice, error) {
	return &virtioRNG{}, nil
}

func (dev *virtioRNG) ToCmdLine() ([]string, error) {
	return []string{"--device", "virtio-rng"}, nil
}

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
