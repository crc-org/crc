package machine

import (
	"encoding/json"
	"errors"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	machineHyperkit "github.com/code-ready/machine/drivers/hyperkit"
	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/host"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(hyperkit.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("hyperkit", constants.CrcBinDir, json)
}

func loadDriverConfig(host *host.Host) (*machineHyperkit.Driver, error) {
	var hyperkitDriver machineHyperkit.Driver
	err := json.Unmarshal(host.RawDriver, &hyperkitDriver)

	return &hyperkitDriver, err
}

func updateDriverConfig(host *host.Host, driver *machineHyperkit.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	return host.UpdateConfig(driverData)
}
