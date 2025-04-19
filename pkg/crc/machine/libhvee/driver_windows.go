package libhvee

import (
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	"github.com/crc-org/crc/v2/pkg/drivers/libhvee"
	"github.com/crc-org/machine/libmachine/drivers"
)

func CreateHost(machineConfig config.MachineConfig) *libhvee.Driver {
	libhveeDriver := libhvee.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, libhveeDriver.VMDriver)

	libhveeDriver.SharedDirs = configureShareDirs(machineConfig)
	return libhveeDriver
}

func configureShareDirs(machineConfig config.MachineConfig) []drivers.SharedDir {
	var sharedDirs []drivers.SharedDir
	for _, dir := range machineConfig.SharedDirs {
		sharedDir := drivers.SharedDir{
			Source:   dir,
			Target:   ConvertToUnixPath(dir),
			Tag:      "crc-dir0", // smb share 'crc-dir0' is created in the msi
			Type:     "cifs",
			Username: machineConfig.SharedDirUsername,
		}
		sharedDir9p := drivers.SharedDir{
			Source: dir,
			Target: ConvertToUnixPath(dir) + "9p", // temporary solution until smb sharing is removed
			Tag:    "crc-dir0",                    // same as above
			//Tag:    fmt.Sprintf("dir%d", i),
			Type: "9p",
		}
		sharedDirs = append(sharedDirs, sharedDir, sharedDir9p)
	}
	return sharedDirs
}
