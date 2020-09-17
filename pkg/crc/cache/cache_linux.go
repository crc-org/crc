package cache

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	crcos "github.com/code-ready/crc/pkg/os"
)

func NewMachineDriverLibvirtCache() *Cache {
	return New(libvirt.MachineDriverCommand, libvirt.MachineDriverDownloadURL, constants.CrcBinDir, libvirt.MachineDriverVersion, getCurrentLibvirtDriverVersion)
}

func getCurrentLibvirtDriverVersion() (string, error) {
	driverBinPath := filepath.Join(constants.CrcBinDir, libvirt.MachineDriverCommand)
	stdOut, _, err := crcos.RunWithDefaultLocale(driverBinPath, "version")
	if len(strings.Split(stdOut, ":")) < 2 {
		return "", fmt.Errorf("Unable to parse the version information of %s", driverBinPath)
	}
	return strings.TrimSpace(strings.Split(stdOut, ":")[1]), err
}
