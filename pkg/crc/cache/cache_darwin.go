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

func NewHyperKitCache() *Cache {
	return New(hyperkit.HyperKitCommand, hyperkit.HyperKitDownloadURL, constants.CrcBinDir, hyperkit.HyperKitVersion, getHyperKitVersion)
}

func getHyperKitMachineDriverVersion(binaryPath string) (string, error) {
	return getVersionGeneric(binaryPath, "version")
}

/* This is very similar to cache.getVersionGeneric, except that it's reading from stderr instead of
 * stdout, and it needs to deal with multiline output
 */
func getHyperKitVersion(binaryPath string) (string, error) {
	_, stderr, err := crcos.RunWithDefaultLocale(binaryPath, "-v")
	if err != nil {
		return "", err
	}
	stderr, err = getFirstLine(stderr)
	if err != nil {
		return "", err
	}
	parsedOutput := strings.Split(stderr, ":")
	if len(parsedOutput) < 2 {
		return "", fmt.Errorf("Unable to parse the version information of %s", binaryPath)
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
