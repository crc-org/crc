package host

import (
	"testing"

	"github.com/code-ready/crc/pkg/drivers/none"
	"github.com/stretchr/testify/assert"
)

func TestMigrateHost(t *testing.T) {
	testCases := []struct {
		description            string
		rawData                []byte
		expectedHostAfter      *Host
		expectedMigrationError error
	}{
		{
			// Point of this test is largely that no matter what was in RawDriver
			// before, it should load into the Host struct based on what is actually
			// in the Driver field.
			//
			// Note that we don't check for the presence of RawDriver's literal "on
			// disk" here.  It's intentional.
			description: "Config version 3 load with existing RawDriver on disk",
			rawData: []byte(`{
    "ConfigVersion": 3,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.code-ready/machine/machines/default"
        }
    },
    "Name": "default",
    "RawDriver": "eyJWQm94TWFuYWdlciI6e30sIklQQWRkcmVzcyI6IjE5Mi4xNjguOTkuMTAwIiwiTWFjaGluZU5hbWUiOiJkZWZhdWx0IiwiU1NIVXNlciI6ImRvY2tlciIsIlNTSFBvcnQiOjU4MTQ1LCJTU0hLZXlQYXRoIjoiL1VzZXJzL25hdGhhbmxlY2xhaXJlLy5kb2NrZXIvbWFjaGluZS9tYWNoaW5lcy9kZWZhdWx0L2lkX3JzYSIsIlN0b3JlUGF0aCI6Ii9Vc2Vycy9uYXRoYW5sZWNsYWlyZS8uZG9ja2VyL21hY2hpbmUiLCJTd2FybU1hc3RlciI6ZmFsc2UsIlN3YXJtSG9zdCI6InRjcDovLzAuMC4wLjA6MzM3NiIsIlN3YXJtRGlzY292ZXJ5IjoiIiwiQ1BVIjoxLCJNZW1vcnkiOjEwMjQsIkRpc2tTaXplIjoyMDAwMCwiQm9vdDJEb2NrZXJVUkwiOiIiLCJCb290MkRvY2tlckltcG9ydFZNIjoiIiwiSG9zdE9ubHlDSURSIjoiMTkyLjE2OC45OS4xLzI0IiwiSG9zdE9ubHlOaWNUeXBlIjoiODI1NDBFTSIsIkhvc3RPbmx5UHJvbWlzY01vZGUiOiJkZW55IiwiTm9TaGFyZSI6ZmFsc2V9"
}`),
			expectedHostAfter: &Host{
				ConfigVersion: 3,
				Name:          "default",
				DriverName:    "virtualbox",
				RawDriver:     []byte(`{"MachineName": "default"}`),
				Driver: &RawDataDriver{
					Data: []byte(`{"MachineName": "default"}`),

					// TODO (nathanleclaire): The "." argument here is a already existing
					// bug (or at least likely to cause them in the future) and most
					// likely should be "/Users/nathanleclaire/.code-ready/machine"
					//
					// These default StorePath settings get over-written when we
					// instantiate the plugin driver, but this seems entirely incidental.
					Driver: none.NewDriver("default", ""),
				},
			},
			expectedMigrationError: nil,
		},
		{
			description: "Config version 4 (from the FUTURE) on disk",
			rawData: []byte(`{
    "ConfigVersion": 4,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.code-ready/machine/machines/default"
        }
    },
    "Name": "default"
}`),
			expectedHostAfter:      nil,
			expectedMigrationError: errUnexpectedConfigVersion,
		},
		{
			description: "Config version 3 load WITHOUT any existing RawDriver field on disk",
			rawData: []byte(`{
    "ConfigVersion": 3,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.code-ready/machine/machines/default"
        }
    },
    "Name": "default"
}`),
			expectedHostAfter: &Host{
				ConfigVersion: 3,
				Name:          "default",
				DriverName:    "virtualbox",
				RawDriver:     []byte(`{"MachineName": "default"}`),
				Driver: &RawDataDriver{
					Data: []byte(`{"MachineName": "default"}`),

					// TODO: See note above.
					Driver: none.NewDriver("default", ""),
				},
			},
			expectedMigrationError: nil,
		},
	}

	for _, tc := range testCases {
		actualHostAfter, actualMigrationError := MigrateHost("default", tc.rawData)

		assert.Equal(t, tc.expectedHostAfter, actualHostAfter)
		assert.Equal(t, tc.expectedMigrationError, actualMigrationError)
	}
}
