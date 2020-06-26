package client

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/libmachine/drivers/plugin/localbinary"
)

// StartDriver starts the desired machine driver if necessary.
func StartDriver() {
	if os.Getenv(localbinary.PluginEnvKey) == localbinary.PluginEnvVal {
		errors.ExitWithMessage(1, "Unregistered driver: %s\n", os.Getenv(localbinary.PluginEnvDriverName))
		return
	}
	localbinary.CurrentBinaryIsCRCMachine = true
}
