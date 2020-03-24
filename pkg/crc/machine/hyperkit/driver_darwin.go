package hyperkit

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/machine/libmachine/drivers"
)

type hyperkitDriver struct {
	*drivers.BaseDriver
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
	HyperKitPath  string
}

func CreateHost(config config.MachineConfig) *hyperkitDriver {
	d := &hyperkitDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: config.Name,
			StorePath:   constants.MachineBaseDir,
			SSHUser:     constants.DefaultSSHUser,
			BundleName:  config.BundleName,
		},
		Memory:      config.Memory,
		CPU:         config.CPUs,
		UUID:        "c3d68012-0208-11ea-9fd7-f2189899ab08",
		Cmdline:     config.KernelCmdLine,
		VmlinuzPath: config.Kernel,
		InitrdPath:  config.Initramfs,
		DiskPathUrl: config.DiskPathURL,
		SSHKeyPath:  config.SSHKeyPath,
	}
	d.HyperKitPath = filepath.Join(constants.CrcBinDir, "hyperkit")

	return d
}
