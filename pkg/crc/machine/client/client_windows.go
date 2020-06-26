package client

import (
	"os"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/drivers/hyperv"
	"github.com/code-ready/crc/pkg/libmachine/drivers/plugin"
	"github.com/code-ready/crc/pkg/libmachine/drivers/plugin/localbinary"
)

// StartDriver starts the desired machine driver if necessary.
func StartDriver() {
	if os.Getenv(localbinary.PluginEnvKey) == localbinary.PluginEnvVal {
		driverName := os.Getenv(localbinary.PluginEnvDriverName)
		switch driverName {
		case "hyperv":
			plugin.RegisterDriver(hyperv.NewDriver("", ""))
		default:
			errors.ExitWithMessage(1, "Unregistered driver: %s\n", driverName)
		}
		return
	}
	localbinary.CurrentBinaryIsCRCMachine = true
}
