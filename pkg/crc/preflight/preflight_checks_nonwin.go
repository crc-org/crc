// +build !windows

package preflight

import (
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"
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
