package machine

import (
	"encoding/json"
	"errors"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/vfkit"
	machineVf "github.com/code-ready/crc/pkg/drivers/vfkit"
	"github.com/code-ready/crc/pkg/libmachine"
	"github.com/code-ready/crc/pkg/libmachine/host"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(vfkit.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("vf", constants.BinDir(), json)
}

func loadDriverConfig(host *host.Host) (*machineVf.Driver, error) {
	var vfDriver machineVf.Driver
	err := json.Unmarshal(host.RawDriver, &vfDriver)

	return &vfDriver, err
}

func updateDriverConfig(host *host.Host, driver *machineVf.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	return host.UpdateConfig(driverData)
}
