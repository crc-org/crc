package fakedriver

import (
	"fmt"

	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	MockState state.State
	MockIP    string
	MockName  string
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "Driver"
}

func (d *Driver) UpdateConfigRaw(rawData []byte) error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MockName
}

func (d *Driver) GetIP() (string, error) {
	if d.MockState == state.Error {
		return "", fmt.Errorf("Unable to get ip")
	}
	if d.MockState == state.Timeout {
		select {} // Loop forever
	}
	if d.MockState != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}
	return d.MockIP, nil
}

func (d *Driver) GetState() (state.State, error) {
	return d.MockState, nil
}

func (d *Driver) Create() error {
	return nil
}

func (d *Driver) Start() error {
	d.MockState = state.Running
	return nil
}

func (d *Driver) Stop() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) Kill() error {
	d.MockState = state.Stopped
	return nil
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Upgrade() error {
	return nil
}
