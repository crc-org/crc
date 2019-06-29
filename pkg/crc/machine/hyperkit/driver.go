package hyperkit

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/pborman/uuid"
	"path/filepath"
)

type hyperkitDriver struct {
	*drivers.BaseDriver
	DiskImage     string
	CPU           int
	Memory        int
	Cmdline       string
	UUID          string
	VpnKitSock    string
	VSockPorts    []string
	VmlinuzPath   string
	InitrdPath    string
	KernelCmdLine string
	DiskPathUrl   string
	SSHKeyPath    string
	Hyperkit      string
}

func CreateHost(config config.MachineConfig) *hyperkitDriver {
	return &hyperkitDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: config.Name,
			StorePath:   constants.MachineBaseDir,
			SSHUser:     constants.DefaultSSHUser,
		},
		Memory:      config.Memory,
		CPU:         config.CPUs,
		UUID:        uuid.NewUUID().String(),
		Cmdline:     config.KernelCmdLine,
		VmlinuzPath: config.Kernel,
		InitrdPath:  config.Initramfs,
		// This force to add entry of DiskPath under crc machine config.json
		DiskImage:   filepath.Join(constants.MachineBaseDir, "machines", config.Name, constants.DefaultName),
		DiskPathUrl: config.DiskPathURL,
		SSHKeyPath:  config.SSHKeyPath,
		Hyperkit:    "/Users/prkumar/bin/hyperkit",
	}
}
