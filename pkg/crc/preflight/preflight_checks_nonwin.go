// +build !windows

package preflight

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/embed"
)

var nonWinPreflightChecks = [...]PreflightCheck{
	{
		configKeySuffix:  "check-root-user",
		checkDescription: "Checking if running as non-root",
		check:            checkIfRunningAsNormalUser,
		fixDescription:   "crc should be ran as a normal user",
		flags:            NoFix,
	},
}

func checkIfRunningAsNormalUser() error {
	if os.Geteuid() != 0 {
		return nil
	}
	logging.Debug("Ran as root")
	return fmt.Errorf("crc should be ran as a normal user")
}

func extractBinary(binaryName string, mode os.FileMode) (string, error) {
	destPath := filepath.Join(constants.CrcBinDir, binaryName)
	err := embed.Extract(binaryName, destPath)
	if err != nil {
		return "", err
	}

	err = os.Chmod(destPath, mode)
	if err != nil {
		os.Remove(destPath)
		return "", err
	}

	return destPath, nil
}
