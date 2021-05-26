package libmachine

import (
	"encoding/json"

	"github.com/code-ready/crc/pkg/drivers/hyperv"
	"github.com/code-ready/crc/pkg/libmachine/host"
)

func (api *Client) NewHost(driverName string, driverPath string, rawDriver []byte) (*host.Host, error) {
	driver := hyperv.NewDriver("", "")
	if err := json.Unmarshal(rawDriver, &driver); err != nil {
		return nil, err
	}

	return &host.Host{
		ConfigVersion: host.Version,
		Name:          driver.GetMachineName(),
		Driver:        driver,
		DriverName:    driver.DriverName(),
		DriverPath:    driverPath,
		RawDriver:     rawDriver,
	}, nil
}

func (api *Client) Load(name string) (*host.Host, error) {
	h, err := api.Filestore.Load(name)
	if err != nil {
		return nil, err
	}

	driver := hyperv.NewDriver("", "")
	if err := json.Unmarshal(h.RawDriver, &driver); err != nil {
		return nil, err
	}
	h.Driver = driver
	return h, nil
}
