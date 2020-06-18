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
		errors.ExitWithMessage(1, fmt.Sprintf("Unregistered driver: %s\n", os.Getenv(localbinary.PluginEnvDriverName)))
		return
	}
	localbinary.CurrentBinaryIsCRCMachine = true
}
