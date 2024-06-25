package machine

import (
	"encoding/json"
	"errors"

	"github.com/crc-org/crc/v2/pkg/crc/machine/config"
	"github.com/crc-org/crc/v2/pkg/crc/machine/libhvee"
	machineLibhvee "github.com/crc-org/crc/v2/pkg/drivers/libhvee"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	"github.com/crc-org/crc/v2/pkg/libmachine/host"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(libhvee.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("hyperv", "", json)
}

func loadDriverConfig(host *host.Host) (*machineLibhvee.Driver, error) {
	var libhveeDriver machineLibhvee.Driver
	err := json.Unmarshal(host.RawDriver, &libhveeDriver)

	return &libhveeDriver, err
}

func updateDriverConfig(host *host.Host, driver *machineLibhvee.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}
	return host.UpdateConfig(driverData)
}

func updateKernelArgs(_ *virtualMachine) error {
	return nil
}

func updateDriverStruct(host *host.Host, driver *machineLibhvee.Driver) error {
	host.Driver = driver
	return nil
}
