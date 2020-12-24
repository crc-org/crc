package machine

import (
	"encoding/json"
	"errors"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/libmachine"
	"github.com/code-ready/crc/pkg/libmachine/host"
	machineLibvirt "github.com/code-ready/machine/drivers/libvirt"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(libvirt.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("libvirt", constants.CrcBinDir, json)
}

/* FIXME: host.Host is only known here, and libvirt.Driver is only accessible
 * in libvirt/driver_linux.go
 */
func loadDriverConfig(host *host.Host) (*machineLibvirt.Driver, error) {
	var libvirtDriver machineLibvirt.Driver
	err := json.Unmarshal(host.RawDriver, &libvirtDriver)

	return &libvirtDriver, err
}

func updateDriverConfig(host *host.Host, driver *machineLibvirt.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}
	return host.UpdateConfig(driverData)
}

/*
func (r *RPCServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	return json.Unmarshal(data, &r.ActualDriver)
}
*/
