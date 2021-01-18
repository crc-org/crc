package host

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/code-ready/crc/pkg/drivers/none"
)

var (
	errUnexpectedConfigVersion = errors.New("unexpected config version")
)

type RawDataDriver struct {
	*none.Driver
	Data []byte // passed directly back when invoking json.Marshal on this type
}

func (r *RawDataDriver) MarshalJSON() ([]byte, error) {
	return r.Data, nil
}

func (r *RawDataDriver) UnmarshalJSON(data []byte) error {
	r.Data = data
	return nil
}

func (r *RawDataDriver) UpdateConfigRaw(rawData []byte) error {
	return r.UnmarshalJSON(rawData)
}

func MigrateHost(name string, data []byte) (*Host, error) {
	var hostMetadata Metadata
	if err := json.Unmarshal(data, &hostMetadata); err != nil {
		return nil, err
	}

	if hostMetadata.ConfigVersion != Version {
		return nil, errUnexpectedConfigVersion
	}

	driver := &RawDataDriver{none.NewDriver(name, ""), nil}
	h := Host{
		Name:   name,
		Driver: driver,
	}
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("Error unmarshalling most recent host version: %s", err)
	}
	h.RawDriver = driver.Data
	return &h, nil
}
