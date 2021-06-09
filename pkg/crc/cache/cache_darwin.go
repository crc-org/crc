package cache

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	crcos "github.com/code-ready/crc/pkg/os"
)

func NewMachineDriverHyperKitCache() *Cache {
	return New(hyperkit.MachineDriverCommand, hyperkit.MachineDriverDownloadURL, constants.CrcBinDir, hyperkit.MachineDriverVersion, getHyperKitMachineDriverVersion)
}

func NewQcowToolCache() *Cache {
	return New(hyperkit.QcowToolCommand, hyperkit.QcowToolDownloadURL, constants.CrcBinDir, hyperkit.QcowToolVersion, getQcowToolVersion)
}

func NewHyperKitCache() *Cache {
	return New(hyperkit.HyperKitCommand, hyperkit.HyperKitDownloadURL, constants.CrcBinDir, hyperkit.HyperKitVersion, getHyperKitVersion)
}

func getHyperKitMachineDriverVersion(executablePath string) (string, error) {
	return getVersionGeneric(executablePath, "version")
}

func getQcowToolVersion(executablePath string) (string, error) {
	stdout, _, err := crcos.RunWithDefaultLocale(executablePath, "--version")
	return strings.TrimSpace(stdout), err
}

/* This is very similar to cache.getVersionGeneric, except that it's reading from stderr instead of
 * stdout, and it needs to deal with multiline output
 */
func getHyperKitVersion(executablePath string) (string, error) {
	_, stderr, err := crcos.RunWithDefaultLocale(executablePath, "-v")
	if err != nil {
		return "", err
	}
	stderr, err = getFirstLine(stderr)
	if err != nil {
		return "", err
	}
	parsedOutput := strings.Split(stderr, ":")
	if len(parsedOutput) < 2 {
		return "", fmt.Errorf("Unable to parse the version information of %s", executablePath)
	}
	return strings.TrimSpace(parsedOutput[1]), err
}

func getFirstLine(s string) (line string, err error) {
	reader := bufio.NewReader(strings.NewReader(s))
	line, err = reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	return line, nil
}
