package machine

import (
	"encoding/json"
	"errors"

	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	"github.com/crc-org/crc/v2/pkg/crc/machine/vfkit"
	machineVf "github.com/crc-org/crc/v2/pkg/drivers/vfkit"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	"github.com/crc-org/crc/v2/pkg/libmachine/host"
	"github.com/crc-org/machine/libmachine/drivers"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(vfkit.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("vf", "", json)
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

func updateDriverStruct(_ *host.Host, _ *machineVf.Driver) error {
	return drivers.ErrNotImplemented
}
