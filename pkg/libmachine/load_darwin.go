package libmachine

import (
	"encoding/json"

	"github.com/crc-org/crc/v2/pkg/drivers/vfkit"
	"github.com/crc-org/crc/v2/pkg/libmachine/host"
)

func (api *Client) NewHost(_ string, driverPath string, rawDriver []byte) (*host.Host, error) {
	driver := vfkit.NewDriver("", "")
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

	driver := vfkit.NewDriver("", "")
	if err := json.Unmarshal(h.RawDriver, &driver); err != nil {
		return nil, err
	}
	h.Driver = driver
	return h, nil
}
