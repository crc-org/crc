package client

import (
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/machine/drivers/hyperv"
	"github.com/code-ready/machine/drivers/virtualbox"
	"github.com/code-ready/machine/libmachine/drivers/plugin"
	"github.com/code-ready/machine/libmachine/drivers/plugin/localbinary"
)

// StartDriver starts the desired machine driver if necessary.
func StartDriver() {
	if os.Getenv(localbinary.PluginEnvKey) == localbinary.PluginEnvVal {
		driverName := os.Getenv(localbinary.PluginEnvDriverName)
		switch driverName {
		case "virtualbox":
			plugin.RegisterDriver(virtualbox.NewDriver("", ""))
		case "hyperv":
			plugin.RegisterDriver(hyperv.NewDriver("", ""))
		default:
			errors.ExitWithMessage(1, fmt.Sprintf("Unregistered driver: %s\n", driverName))
		}
		return
	}
	localbinary.CurrentBinaryIsCRCMachine = true
}
