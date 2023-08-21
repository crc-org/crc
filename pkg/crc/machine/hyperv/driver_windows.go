package hyperv

import (
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/drivers/hyperv"
	winnet "github.com/crc-org/crc/v2/pkg/os/windows/network"
	"github.com/crc-org/machine/libmachine/drivers"
)

func CreateHost(machineConfig config.MachineConfig) *hyperv.Driver {
	hypervDriver := hyperv.NewDriver(machineConfig.Name, constants.MachineBaseDir)

	config.InitVMDriverFromMachineConfig(machineConfig, hypervDriver.VMDriver)

	hypervDriver.DisableDynamicMemory = true

	if machineConfig.NetworkMode == network.UserNetworkingMode {
		hypervDriver.VirtualSwitch = ""
	} else {
		// Determine the Virtual Switch to be used
		_, switchName := winnet.SelectSwitchByNameOrDefault(AlternativeNetwork)
		hypervDriver.VirtualSwitch = switchName
	}

	hypervDriver.SharedDirs = configureShareDirs(machineConfig)
	return hypervDriver
}

// converts a path like c:\users\crc to /mnt/c/users/crc
func convertToUnixPath(path string) string {
	/* podman internally converts windows style paths like C:\Users\crc  to
	 * /mnt/c/Users/crc so it expects the shared folder to be mounted under
	 * '/mnt' instead of '/' like in the case of macOS and linux
	 * see: https://github.com/containers/podman/blob/468aa6478c73e4acd8708ce8bb0bb5a056f329c2/pkg/specgen/winpath.go#L24-L59
	 */
	path = filepath.ToSlash(path)
	if len(path) > 1 && path[1] == ':' {
		return ("/mnt/" + strings.ToLower(path[0:1]) + path[2:])
	}
	return path
}

func configureShareDirs(machineConfig config.MachineConfig) []drivers.SharedDir {
	var sharedDirs []drivers.SharedDir
	for _, dir := range machineConfig.SharedDirs {
		sharedDir := drivers.SharedDir{
			Source:   dir,
			Target:   convertToUnixPath(dir),
			Tag:      "crc-dir0", // smb share 'crc-dir0' is created in the msi
			Type:     "cifs",
			Username: machineConfig.SharedDirUsername,
		}
		sharedDirs = append(sharedDirs, sharedDir)
	}
	return sharedDirs
}
