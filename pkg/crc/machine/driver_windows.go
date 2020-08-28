package machine

import (
	"encoding/json"
	"errors"

	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperv"
	machineHyperv "github.com/code-ready/machine/drivers/hyperv"
	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/host"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(hyperv.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("hyperv", "", json)
}

func loadDriverConfig(host *host.Host) (*machineHyperv.Driver, error) {
	var hypervDriver machineHyperv.Driver
	err := json.Unmarshal(host.RawDriver, &hypervDriver)

	return &hypervDriver, err
}

func updateDriverConfig(host *host.Host, driver *machineHyperv.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}
	return host.UpdateConfig(driverData)
}
