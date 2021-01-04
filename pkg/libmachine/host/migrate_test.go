package host

import (
	"testing"

	"github.com/code-ready/crc/pkg/drivers/none"
	"github.com/stretchr/testify/assert"
)

func TestLoadUnsupportedConfiguration(t *testing.T) {
	_, err := MigrateHost("default", []byte(`{"ConfigVersion": 4}`))
	assert.Equal(t, err, errUnexpectedConfigVersion)
}

func TestLoadHost(t *testing.T) {
	driverJSON := `{
        "IPAddress": "192.168.130.11",
        "MachineName": "crc",
        "BundleName": "crc_libvirt_4.6.6.crcbundle",
        "Memory": 9216,
        "CPU": 4
    }`

	host, err := MigrateHost("default", []byte(`{
    "ConfigVersion": 3,
    "Driver": `+driverJSON+`,
    "DriverName": "libvirt",
    "DriverPath": "/home/john/.crc/bin",
    "Name": "crc"
}`))

	assert.NoError(t, err)
	assert.Equal(t, &Host{
		ConfigVersion: 3,
		Name:          "crc",
		DriverName:    "libvirt",
		DriverPath:    "/home/john/.crc/bin",
		RawDriver:     []byte(driverJSON),
		Driver: &RawDataDriver{
			Data:   []byte(driverJSON),
			Driver: none.NewDriver("default", ""),
		},
	}, host)
}
