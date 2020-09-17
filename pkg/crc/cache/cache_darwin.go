package cache

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	crcos "github.com/code-ready/crc/pkg/os"
)

func NewMachineDriverHyperkitCache() *Cache {
	return New(hyperkit.MachineDriverCommand, hyperkit.MachineDriverDownloadURL, constants.CrcBinDir, hyperkit.MachineDriverVersion, getHyperKitMachineDriverVersion)
}

func NewHyperkitCache() *Cache {
	return New(hyperkit.HyperkitCommand, hyperkit.HyperkitDownloadURL, constants.CrcBinDir, "", nil)
}

func getHyperKitMachineDriverVersion() (string, error) {
	driverBinPath := filepath.Join(constants.CrcBinDir, hyperkit.MachineDriverCommand)
	stdOut, _, err := crcos.RunWithDefaultLocale(driverBinPath, "version")
	if len(strings.Split(stdOut, ":")) < 2 {
		return "", fmt.Errorf("Unable to parse the version information of %s", driverBinPath)
	}
	return strings.TrimSpace(strings.Split(stdOut, ":")[1]), err
}
