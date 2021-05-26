// +build !windows

package libmachine

import (
	"github.com/code-ready/crc/pkg/libmachine/host"
)

func (api *Client) NewHost(driverName string, driverPath string, rawDriver []byte) (*host.Host, error) {
	driver, err := api.clientDriverFactory.NewRPCClientDriver(driverName, driverPath, rawDriver)
	if err != nil {
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

	d, err := api.clientDriverFactory.NewRPCClientDriver(h.DriverName, h.DriverPath, h.RawDriver)
	if err != nil {
		return nil, err
	}
	h.Driver = d
	return h, nil
}
