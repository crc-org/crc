package client

import (
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/machine/libmachine/drivers/plugin/localbinary"
)

// StartDriver starts the desired machine driver if necessary.
func StartDriver() {
	if os.Getenv(localbinary.PluginEnvKey) == localbinary.PluginEnvVal {
		driverName := os.Getenv(localbinary.PluginEnvDriverName)
		switch driverName {
		default:
			errors.ExitWithMessage(1, fmt.Sprintf("Unregistered driver: %s\n", driverName))
		}
		return
	}
	localbinary.CurrentBinaryIsCRCMachine = true
}
