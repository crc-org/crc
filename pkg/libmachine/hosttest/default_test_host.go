package hosttest

import (
	"github.com/code-ready/crc/pkg/drivers/none"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/code-ready/crc/pkg/libmachine/version"
)

const (
	DefaultHostName = "test-host"
)

func GetDefaultTestHost() (*host.Host, error) {
	driver := none.NewDriver(DefaultHostName, "/tmp/artifacts")

	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          DefaultHostName,
		Driver:        driver,
		DriverName:    "none",
	}, nil
}
