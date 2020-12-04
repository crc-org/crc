package drivers

import (
	"errors"
	"path/filepath"
)

// BaseDriver - Embed this struct into drivers to provide the common set
// of fields and functions.
type BaseDriver struct {
	IPAddress   string
	MachineName string
	StorePath   string
	BundleName  string
}

type VMDriver struct {
	*BaseDriver
	ImageSourcePath string
	ImageFormat     string
	Memory          int
	CPU             int
	DiskCapacity    uint64 // bytes
}

// DriverName returns the name of the driver
func (d *BaseDriver) DriverName() string {
	return "unknown"
}

// DriverName returns the name of the driver
func (d *BaseDriver) DriverVersion() string {
	return "unknown"
}

// GetMachineName returns the machine name
func (d *BaseDriver) GetMachineName() string {
	return d.MachineName
}

// GetIP returns the ip
func (d *BaseDriver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", errors.New("IP address is not set")
	}
	return d.IPAddress, nil
}

// PreCreateCheck is called to enforce pre-creation steps
func (d *BaseDriver) PreCreateCheck() error {
	return nil
}

// ResolveStorePath returns the store path where the machine is
func (d *BaseDriver) ResolveStorePath(file string) string {
	return filepath.Join(d.StorePath, "machines", d.MachineName, file)
}

// Returns the name of the bundle which was used to create this machine
func (d *BaseDriver) GetBundleName() (string, error) {
	if d.BundleName == "" {
		return "", errors.New("Bundle name is not set")
	}
	return d.BundleName, nil
}

func (d *BaseDriver) UpdateConfigRaw(rawData []byte) error {
	return ErrNotImplemented
}
