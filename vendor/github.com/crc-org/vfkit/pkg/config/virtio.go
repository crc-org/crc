package config

import (
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// The VirtioDevice interface is an interface which is implemented by all virtio devices.
type VirtioDevice VMComponent

const (
	// Possible values for VirtioInput.InputType
	VirtioInputPointingDevice = "pointing"
	VirtioInputKeyboardDevice = "keyboard"

	// Options for VirtioGPUResolution
	VirtioGPUResolutionWidth  = "width"
	VirtioGPUResolutionHeight = "height"

	// Default VirtioGPU Resolution
	defaultVirtioGPUResolutionWidth  = 800
	defaultVirtioGPUResolutionHeight = 600
)

// VirtioInput configures an input device, such as a keyboard or pointing device
// (mouse) that the virtual machine can use
type VirtioInput struct {
	InputType string `json:"inputType"` // currently supports "pointing" and "keyboard" input types
}

type VirtioGPUResolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// VirtioGPU configures a GPU device, such as the host computer's display
type VirtioGPU struct {
	UsesGUI bool `json:"usesGUI"`
	VirtioGPUResolution
}

// VirtioVsock configures of a virtio-vsock device allowing 2-way communication
// between the host and the virtual machine type
type VirtioVsock struct {
	// Port is the virtio-vsock port used for this device, see `man vsock` for more
	// details.
	Port uint32 `json:"port"`
	// SocketURL is the path to a unix socket on the host to use for the virtio-vsock communication with the guest.
	SocketURL string `json:"socketURL"`
	// If true, vsock connections will have to be done from guest to host. If false, vsock connections will only be possible
	// from host to guest
	Listen bool `json:"listen,omitempty"`
}

// VirtioBlk configures a disk device.
type VirtioBlk struct {
	StorageConfig
	DeviceIdentifier string `json:"deviceIdentifier,omitempty"`
}

type DirectorySharingConfig struct {
	MountTag string `json:"mountTag"`
}

// VirtioFs configures directory sharing between the guest and the host.
type VirtioFs struct {
	DirectorySharingConfig
	SharedDir string `json:"sharedDir"`
}

// RosettaShare configures rosetta support in the guest to run Intel binaries on Apple CPUs
type RosettaShare struct {
	DirectorySharingConfig
	InstallRosetta  bool `json:"installRosetta"`
	IgnoreIfMissing bool `json:"ignoreIfMissing"`
}

// NVMExpressController configures a NVMe controller in the guest
type NVMExpressController struct {
	StorageConfig
}

// VirtioRng configures a random number generator (RNG) device.
type VirtioRng struct {
}

// TODO: Add BridgedNetwork support
// https://github.com/Code-Hex/vz/blob/d70a0533bf8ed0fa9ab22fa4d4ca554b7c3f3ce5/network.go#L81-L82

// VirtioNet configures the virtual machine networking.
type VirtioNet struct {
	Nat        bool             `json:"nat"`
	MacAddress net.HardwareAddr `json:"-"` // custom marshaller in json.go
	// file parameter is holding a connected datagram socket.
	// see https://github.com/Code-Hex/vz/blob/7f648b6fb9205d6f11792263d79876e3042c33ec/network.go#L113-L155
	Socket *os.File `json:"socket,omitempty"`

	UnixSocketPath string `json:"unixSocketPath,omitempty"`
}

// VirtioSerial configures the virtual machine serial ports.
type VirtioSerial struct {
	LogFile   string `json:"logFile,omitempty"`
	UsesStdio bool   `json:"usesStdio,omitempty"`
}

type NBDSynchronizationMode string

const (
	SynchronizationFullMode NBDSynchronizationMode = "full"
	SynchronizationNoneMode NBDSynchronizationMode = "none"
)

type NetworkBlockDevice struct {
	VirtioBlk
	Timeout             time.Duration
	SynchronizationMode NBDSynchronizationMode
}

// TODO: Add VirtioBalloon
// https://github.com/Code-Hex/vz/blob/master/memory_balloon.go

type option struct {
	key   string
	value string
}

func strToOption(str string) option {
	splitStr := strings.SplitN(str, "=", 2)

	opt := option{
		key: splitStr[0],
	}
	if len(splitStr) > 1 {
		opt.value = splitStr[1]
	}

	return opt
}

func strvToOptions(opts []string) []option {
	parsedOpts := []option{}
	for _, opt := range opts {
		if len(opt) == 0 {
			continue
		}
		parsedOpts = append(parsedOpts, strToOption(opt))
	}

	return parsedOpts
}

func deviceFromCmdLine(deviceOpts string) (VirtioDevice, error) {
	opts := strings.Split(deviceOpts, ",")
	if len(opts) == 0 {
		return nil, fmt.Errorf("empty option list in command line argument")
	}
	var dev VirtioDevice
	switch opts[0] {
	case "rosetta":
		dev = &RosettaShare{}
	case "nvme":
		dev = nvmExpressControllerNewEmpty()
	case "virtio-blk":
		dev = virtioBlkNewEmpty()
	case "virtio-fs":
		dev = &VirtioFs{}
	case "virtio-net":
		dev = &VirtioNet{}
	case "virtio-rng":
		dev = &VirtioRng{}
	case "virtio-serial":
		dev = &VirtioSerial{}
	case "virtio-vsock":
		dev = &VirtioVsock{}
	case "usb-mass-storage":
		dev = usbMassStorageNewEmpty()
	case "virtio-input":
		dev = &VirtioInput{}
	case "virtio-gpu":
		dev = &VirtioGPU{}
	case "nbd":
		dev = networkBlockDeviceNewEmpty()
	default:
		return nil, fmt.Errorf("unknown device type: %s", opts[0])
	}

	parsedOpts := strvToOptions(opts[1:])
	if err := dev.FromOptions(parsedOpts); err != nil {
		return nil, err
	}

	return dev, nil
}

// VirtioSerialNew creates a new serial device for the virtual machine. The
// output the virtual machine sent to the serial port will be written to the
// file at logFilePath.
func VirtioSerialNew(logFilePath string) (VirtioDevice, error) {
	return &VirtioSerial{
		LogFile: logFilePath,
	}, nil
}

func VirtioSerialNewStdio() (VirtioDevice, error) {
	return &VirtioSerial{
		UsesStdio: true,
	}, nil
}

func (dev *VirtioSerial) validate() error {
	if dev.LogFile != "" && dev.UsesStdio {
		return fmt.Errorf("'logFilePath' and 'stdio' cannot be set at the same time")
	}
	if dev.LogFile == "" && !dev.UsesStdio {
		return fmt.Errorf("one of 'logFilePath' or 'stdio' must be set")
	}

	return nil
}

func (dev *VirtioSerial) ToCmdLine() ([]string, error) {
	if err := dev.validate(); err != nil {
		return nil, err
	}
	if dev.UsesStdio {
		return []string{"--device", "virtio-serial,stdio"}, nil
	}

	return []string{"--device", fmt.Sprintf("virtio-serial,logFilePath=%s", dev.LogFile)}, nil
}

func (dev *VirtioSerial) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "logFilePath":
			dev.LogFile = option.value
		case "stdio":
			if option.value != "" {
				return fmt.Errorf("unexpected value for virtio-serial 'stdio' option: %s", option.value)
			}
			dev.UsesStdio = true
		default:
			return fmt.Errorf("unknown option for virtio-serial devices: %s", option.key)
		}
	}

	return dev.validate()
}

// VirtioInputNew creates a new input device for the virtual machine.
// The inputType parameter is the type of virtio-input device that will be added
// to the machine.
func VirtioInputNew(inputType string) (VirtioDevice, error) {
	dev := &VirtioInput{
		InputType: inputType,
	}
	if err := dev.validate(); err != nil {
		return nil, err
	}

	return dev, nil
}

func (dev *VirtioInput) validate() error {
	if dev.InputType != VirtioInputPointingDevice && dev.InputType != VirtioInputKeyboardDevice {
		return fmt.Errorf("unknown option for virtio-input devices: %s", dev.InputType)
	}

	return nil
}

func (dev *VirtioInput) ToCmdLine() ([]string, error) {
	if err := dev.validate(); err != nil {
		return nil, err
	}

	return []string{"--device", fmt.Sprintf("virtio-input,%s", dev.InputType)}, nil
}

func (dev *VirtioInput) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case VirtioInputPointingDevice, VirtioInputKeyboardDevice:
			if option.value != "" {
				return fmt.Errorf("unexpected value for virtio-input %s option: %s", option.key, option.value)
			}
			dev.InputType = option.key
		default:
			return fmt.Errorf("unknown option for virtio-input devices: %s", option.key)
		}
	}
	return dev.validate()
}

// VirtioGPUNew creates a new gpu device for the virtual machine.
// The usesGUI parameter determines whether a graphical application window will
// be displayed
func VirtioGPUNew() (VirtioDevice, error) {
	return &VirtioGPU{
		UsesGUI: false,
		VirtioGPUResolution: VirtioGPUResolution{
			Width:  defaultVirtioGPUResolutionWidth,
			Height: defaultVirtioGPUResolutionHeight,
		},
	}, nil
}

func (dev *VirtioGPU) validate() error {
	if dev.Height < 1 || dev.Width < 1 {
		return fmt.Errorf("invalid dimensions for virtio-gpu device resolution: %dx%d", dev.Width, dev.Height)
	}

	return nil
}

func (dev *VirtioGPU) ToCmdLine() ([]string, error) {
	if err := dev.validate(); err != nil {
		return nil, err
	}

	return []string{"--device", fmt.Sprintf("virtio-gpu,width=%d,height=%d", dev.Width, dev.Height)}, nil
}

func (dev *VirtioGPU) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case VirtioGPUResolutionHeight:
			height, err := strconv.Atoi(option.value)
			if err != nil || height < 1 {
				return fmt.Errorf("Invalid value for virtio-gpu %s: %s", option.key, option.value)
			}

			dev.Height = height
		case VirtioGPUResolutionWidth:
			width, err := strconv.Atoi(option.value)
			if err != nil || width < 1 {
				return fmt.Errorf("Invalid value for virtio-gpu %s: %s", option.key, option.value)
			}

			dev.Width = width
		default:
			return fmt.Errorf("unknown option for virtio-gpu devices: %s", option.key)
		}
	}

	if dev.Width == 0 && dev.Height == 0 {
		dev.Width = defaultVirtioGPUResolutionWidth
		dev.Height = defaultVirtioGPUResolutionHeight
	}

	return dev.validate()
}

// VirtioNetNew creates a new network device for the virtual machine. It will
// use macAddress as its MAC address.
func VirtioNetNew(macAddress string) (*VirtioNet, error) {
	var hwAddr net.HardwareAddr

	if macAddress != "" {
		var err error
		if hwAddr, err = net.ParseMAC(macAddress); err != nil {
			return nil, err
		}
	}
	return &VirtioNet{
		Nat:        true,
		MacAddress: hwAddr,
	}, nil
}

// SetSocket Set the socket to use for the network communication
//
// This maps the virtual machine network interface to a connected datagram
// socket. This means all network traffic on this interface will go through
// file.
// file must be a connected datagram (SOCK_DGRAM) socket.
func (dev *VirtioNet) SetSocket(file *os.File) {
	dev.Socket = file
	dev.Nat = false
}

func (dev *VirtioNet) SetUnixSocketPath(path string) {
	dev.UnixSocketPath = path
	dev.Nat = false
}

func (dev *VirtioNet) validate() error {
	if dev.Nat && dev.Socket != nil {
		return fmt.Errorf("'nat' and 'fd' cannot be set at the same time")
	}
	if dev.Nat && dev.UnixSocketPath != "" {
		return fmt.Errorf("'nat' and 'unixSocketPath' cannot be set at the same time")
	}
	if dev.Socket != nil && dev.UnixSocketPath != "" {
		return fmt.Errorf("'fd' and 'unixSocketPath' cannot be set at the same time")
	}
	if !dev.Nat && dev.Socket == nil && dev.UnixSocketPath == "" {
		return fmt.Errorf("one of 'nat' or 'fd' or 'unixSocketPath' must be set")
	}

	return nil
}

func (dev *VirtioNet) ToCmdLine() ([]string, error) {
	if err := dev.validate(); err != nil {
		return nil, err
	}

	builder := strings.Builder{}
	builder.WriteString("virtio-net")
	switch {
	case dev.Nat:
		builder.WriteString(",nat")
	case dev.UnixSocketPath != "":
		fmt.Fprintf(&builder, ",unixSocketPath=%s", dev.UnixSocketPath)
	default:
		fmt.Fprintf(&builder, ",fd=%d", dev.Socket.Fd())
	}

	if len(dev.MacAddress) != 0 {
		builder.WriteString(fmt.Sprintf(",mac=%s", dev.MacAddress))
	}

	return []string{"--device", builder.String()}, nil
}

func (dev *VirtioNet) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "nat":
			if option.value != "" {
				return fmt.Errorf("unexpected value for virtio-net 'nat' option: %s", option.value)
			}
			dev.Nat = true
		case "mac":
			macAddress, err := net.ParseMAC(option.value)
			if err != nil {
				return err
			}
			dev.MacAddress = macAddress
		case "fd":
			fd, err := strconv.Atoi(option.value)
			if err != nil {
				return err
			}
			dev.Socket = os.NewFile(uintptr(fd), "vfkit virtio-net socket")
		case "unixSocketPath":
			dev.UnixSocketPath = option.value
		default:
			return fmt.Errorf("unknown option for virtio-net devices: %s", option.key)
		}
	}

	return dev.validate()
}

// VirtioRngNew creates a new random number generator device to feed entropy
// into the virtual machine.
func VirtioRngNew() (VirtioDevice, error) {
	return &VirtioRng{}, nil
}

func (dev *VirtioRng) ToCmdLine() ([]string, error) {
	return []string{"--device", "virtio-rng"}, nil
}

func (dev *VirtioRng) FromOptions(options []option) error {
	if len(options) != 0 {
		return fmt.Errorf("unknown options for virtio-rng devices: %s", options)
	}
	return nil
}

func nvmExpressControllerNewEmpty() *NVMExpressController {
	return &NVMExpressController{
		StorageConfig: StorageConfig{
			DevName: "nvme",
		},
	}
}

// NVMExpressControllerNew creates a new NVMExpress controller to use in the
// virtual machine. It will use the file at imagePath as the disk image. This
// image must be in raw format.
func NVMExpressControllerNew(imagePath string) (*NVMExpressController, error) {
	r := nvmExpressControllerNewEmpty()
	r.ImagePath = imagePath
	return r, nil
}

func virtioBlkNewEmpty() *VirtioBlk {
	return &VirtioBlk{
		StorageConfig: StorageConfig{
			DevName: "virtio-blk",
		},
		DeviceIdentifier: "",
	}
}

// VirtioBlkNew creates a new disk to use in the virtual machine. It will use
// the file at imagePath as the disk image. This image must be in raw format.
func VirtioBlkNew(imagePath string) (*VirtioBlk, error) {
	virtioBlk := virtioBlkNewEmpty()
	virtioBlk.ImagePath = imagePath

	return virtioBlk, nil
}

func (dev *VirtioBlk) SetDeviceIdentifier(devID string) {
	dev.DeviceIdentifier = devID
}

func (dev *VirtioBlk) FromOptions(options []option) error {
	unhandledOpts := []option{}
	for _, option := range options {
		switch option.key {
		case "deviceId":
			dev.DeviceIdentifier = option.value
		default:
			unhandledOpts = append(unhandledOpts, option)
		}
	}

	return dev.StorageConfig.FromOptions(unhandledOpts)
}

func (dev *VirtioBlk) ToCmdLine() ([]string, error) {
	cmdLine, err := dev.StorageConfig.ToCmdLine()
	if err != nil {
		return []string{}, err
	}
	if len(cmdLine) != 2 {
		return []string{}, fmt.Errorf("unexpected storage config commandline")
	}
	if dev.DeviceIdentifier != "" {
		cmdLine[1] = fmt.Sprintf("%s,deviceId=%s", cmdLine[1], dev.DeviceIdentifier)
	}
	return cmdLine, nil
}

// VirtioVsockNew creates a new virtio-vsock device for 2-way communication
// between the host and the virtual machine. The communication will happen on
// vsock port, and on the host it will use the unix socket at socketURL.
// When listen is true, the host will be listening for connections over vsock.
// When listen  is false, the guest will be listening for connections over vsock.
func VirtioVsockNew(port uint, socketURL string, listen bool) (VirtioDevice, error) {
	if port > math.MaxUint32 {
		return nil, fmt.Errorf("invalid vsock port: %d", port)
	}
	return &VirtioVsock{
		Port:      uint32(port), //#nosec G115 -- was compared to math.MaxUint32
		SocketURL: socketURL,
		Listen:    listen,
	}, nil
}

func (dev *VirtioVsock) ToCmdLine() ([]string, error) {
	if dev.Port == 0 || dev.SocketURL == "" {
		return nil, fmt.Errorf("virtio-vsock needs both a port and a socket URL")
	}
	var listenStr string
	if dev.Listen {
		listenStr = "listen"
	} else {
		listenStr = "connect"
	}
	return []string{"--device", fmt.Sprintf("virtio-vsock,port=%d,socketURL=%s,%s", dev.Port, dev.SocketURL, listenStr)}, nil
}

func (dev *VirtioVsock) FromOptions(options []option) error {
	// default to listen for backwards compatibliity
	dev.Listen = true
	for _, option := range options {
		switch option.key {
		case "socketURL":
			dev.SocketURL = option.value
		case "port":
			port, err := strconv.ParseUint(option.value, 10, 32)
			if err != nil {
				return err
			}
			dev.Port = uint32(port) //#nosec G115 -- ParseUint(_, _, 32) guarantees no overflow
		case "listen":
			dev.Listen = true
		case "connect":
			dev.Listen = false
		default:
			return fmt.Errorf("unknown option for virtio-vsock devices: %s", option.key)
		}
	}
	return nil
}

// VirtioFsNew creates a new virtio-fs device for file sharing. It will share
// the directory at sharedDir with the virtual machine. This directory can be
// mounted in the VM using `mount -t virtiofs mountTag /some/dir`
func VirtioFsNew(sharedDir string, mountTag string) (VirtioDevice, error) {
	return &VirtioFs{
		DirectorySharingConfig: DirectorySharingConfig{
			MountTag: mountTag,
		},
		SharedDir: sharedDir,
	}, nil
}

func (dev *VirtioFs) ToCmdLine() ([]string, error) {
	if dev.SharedDir == "" {
		return nil, fmt.Errorf("virtio-fs needs the path to the directory to share")
	}
	if dev.MountTag != "" {
		return []string{"--device", fmt.Sprintf("virtio-fs,sharedDir=%s,mountTag=%s", dev.SharedDir, dev.MountTag)}, nil
	}

	return []string{"--device", fmt.Sprintf("virtio-fs,sharedDir=%s", dev.SharedDir)}, nil
}

func (dev *VirtioFs) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "sharedDir":
			dev.SharedDir = option.value
		case "mountTag":
			dev.MountTag = option.value
		default:
			return fmt.Errorf("unknown option for virtio-fs devices: %s", option.key)
		}
	}
	return nil
}

// RosettaShareNew RosettaShare creates a new rosetta share for running x86_64 binaries on M1 machines.
// It will share a directory containing the linux rosetta binaries with the
// virtual machine. This directory can be mounted in the VM using `mount -t
// virtiofs mountTag /some/dir`
func RosettaShareNew(mountTag string) (VirtioDevice, error) {
	return &RosettaShare{
		DirectorySharingConfig: DirectorySharingConfig{
			MountTag: mountTag,
		},
	}, nil
}

func (dev *RosettaShare) ToCmdLine() ([]string, error) {
	if dev.MountTag == "" {
		return nil, fmt.Errorf("rosetta shares require a mount tag to be specified")
	}
	builder := strings.Builder{}
	builder.WriteString("rosetta")
	fmt.Fprintf(&builder, ",mountTag=%s", dev.MountTag)
	if dev.InstallRosetta {
		builder.WriteString(",install")
	}
	if dev.IgnoreIfMissing {
		builder.WriteString(",ignore-if-missing")
	}

	return []string{"--device", builder.String()}, nil
}

func (dev *RosettaShare) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "mountTag":
			dev.MountTag = option.value
		case "install":
			dev.InstallRosetta = true
		case "ignore-if-missing":
			dev.IgnoreIfMissing = true
		default:
			return fmt.Errorf("unknown option for rosetta share: %s", option.key)
		}
	}
	return nil
}

func networkBlockDeviceNewEmpty() *NetworkBlockDevice {
	return &NetworkBlockDevice{
		VirtioBlk: VirtioBlk{
			StorageConfig: StorageConfig{
				DevName: "nbd",
			},
			DeviceIdentifier: "",
		},
		Timeout:             time.Duration(15000 * time.Millisecond), // set a default timeout to 15s
		SynchronizationMode: SynchronizationFullMode,                 // default mode to full
	}
}

// NetworkBlockDeviceNew creates a new disk by connecting to a remote Network Block Device (NBD) server.
// The provided uri must be in the format <scheme>://<address>/<export-name>
// where scheme could have any of these value: nbd, nbds, nbd+unix and nbds+unix.
// More info can be found at https://github.com/NetworkBlockDevice/nbd/blob/master/doc/uri.md
// This allows the virtual machine to access and use the remote storage as if it were a local disk.
func NetworkBlockDeviceNew(uri string, timeout uint32, synchronization NBDSynchronizationMode) (*NetworkBlockDevice, error) {
	nbd := networkBlockDeviceNewEmpty()
	nbd.URI = uri
	nbd.Timeout = time.Duration(timeout) * time.Millisecond
	nbd.SynchronizationMode = synchronization

	return nbd, nil
}

func (nbd *NetworkBlockDevice) ToCmdLine() ([]string, error) {
	cmdLine, err := nbd.VirtioBlk.ToCmdLine()
	if err != nil {
		return []string{}, err
	}
	if len(cmdLine) != 2 {
		return []string{}, fmt.Errorf("unexpected storage config commandline")
	}

	if nbd.Timeout.Milliseconds() > 0 {
		cmdLine[1] = fmt.Sprintf("%s,timeout=%d", cmdLine[1], nbd.Timeout.Milliseconds())
	}
	if nbd.SynchronizationMode == "none" || nbd.SynchronizationMode == "full" {
		cmdLine[1] = fmt.Sprintf("%s,sync=%s", cmdLine[1], nbd.SynchronizationMode)
	}

	return cmdLine, nil
}

func (nbd *NetworkBlockDevice) FromOptions(options []option) error {
	unhandledOpts := []option{}
	for _, option := range options {
		switch option.key {
		case "timeout":
			timeoutMS, err := strconv.ParseInt(option.value, 10, 32)
			if err != nil {
				return err
			}
			nbd.Timeout = time.Duration(timeoutMS) * time.Millisecond
		case "sync":
			switch option.value {
			case string(SynchronizationFullMode):
				nbd.SynchronizationMode = SynchronizationFullMode
			case string(SynchronizationNoneMode):
				nbd.SynchronizationMode = SynchronizationNoneMode
			default:
				return fmt.Errorf("invalid sync mode: %s, must be 'full' or 'none'", option.value)
			}
		default:
			unhandledOpts = append(unhandledOpts, option)
		}
	}

	return nbd.VirtioBlk.FromOptions(unhandledOpts)
}

type USBMassStorage struct {
	StorageConfig
}

func usbMassStorageNewEmpty() *USBMassStorage {
	return &USBMassStorage{
		StorageConfig{
			DevName: "usb-mass-storage",
		},
	}
}

// USBMassStorageNew creates a new USB disk to use in the virtual machine. It will use
// the file at imagePath as the disk image. This image must be in raw or ISO format.
func USBMassStorageNew(imagePath string) (*USBMassStorage, error) {
	usbMassStorage := usbMassStorageNewEmpty()
	usbMassStorage.ImagePath = imagePath

	return usbMassStorage, nil
}

func (dev *USBMassStorage) SetReadOnly(readOnly bool) {
	dev.StorageConfig.ReadOnly = readOnly
}

// StorageConfig configures a disk device.
type StorageConfig struct {
	DevName   string `json:"devName"`
	ImagePath string `json:"imagePath,omitempty"`
	URI       string `json:"uri,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

func (config *StorageConfig) ToCmdLine() ([]string, error) {
	if config.ImagePath != "" && config.URI != "" {
		return nil, fmt.Errorf("%s devices cannot have both path to a disk image and a uri to a remote block device", config.DevName)
	}
	if config.ImagePath == "" && config.URI == "" {
		return nil, fmt.Errorf("%s devices need a path to a disk image or a uri to a remote block device", config.DevName)
	}
	var value string
	if config.ImagePath != "" {
		value = fmt.Sprintf("%s,path=%s", config.DevName, config.ImagePath)
	}
	if config.URI != "" {
		value = fmt.Sprintf("%s,uri=%s", config.DevName, config.URI)
	}

	if config.ReadOnly {
		value += ",readonly"
	}
	return []string{"--device", value}, nil
}

func (config *StorageConfig) FromOptions(options []option) error {
	for _, option := range options {
		switch option.key {
		case "path":
			config.ImagePath = option.value
		case "uri":
			config.URI = option.value
		case "readonly":
			if option.value != "" {
				return fmt.Errorf("unexpected value for virtio-blk 'readonly' option: %s", option.value)
			}
			config.ReadOnly = true
		default:
			return fmt.Errorf("unknown option for %s devices: %s", config.DevName, option.key)
		}
	}
	return nil
}
