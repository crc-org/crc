package systemd

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/constants"

	"github.com/code-ready/crc/pkg/crc/systemd/actions"

	"github.com/code-ready/machine/libmachine/drivers"
)

type InstanceSystemdCommander struct {
	driver drivers.Driver
}

// NewVmSystemdCommander creates a new instance of a VmSystemdCommander
func NewInstanceSystemdCommander(driver drivers.Driver) *InstanceSystemdCommander {
	return &InstanceSystemdCommander{
		driver: driver,
	}
}

func (c InstanceSystemdCommander) Enable(name string) (bool, error) {
	_, err := c.service(name, actions.Enable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c InstanceSystemdCommander) Disable(name string) (bool, error) {
	_, err := c.service(name, actions.Disable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c InstanceSystemdCommander) DaemonReload() (bool, error) {
	// Might be needed for Start or Restart
	_, err := drivers.RunSSHCommandFromDriver(c.driver, constants.GetPrivateKeyPath(), "sudo systemctl daemon-reload")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c InstanceSystemdCommander) Restart(name string) (bool, error) {
	_, _ = c.DaemonReload()
	_, err := c.service(name, actions.Restart)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c InstanceSystemdCommander) Start(name string) (bool, error) {
	_, _ = c.DaemonReload()
	_, err := c.service(name, actions.Start)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c InstanceSystemdCommander) Stop(name string) (bool, error) {
	_, err := c.service(name, actions.Stop)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c InstanceSystemdCommander) Status(name string) (string, error) {
	return c.service(name, actions.Status)

}

func (c InstanceSystemdCommander) service(name string, action actions.Action) (string, error) {
	command := fmt.Sprintf("sudo systemctl -f %s %s", action.String(), name)

	if out, err := drivers.RunSSHCommandFromDriver(c.driver, constants.GetPrivateKeyPath(), command); err != nil {
		return out, err
	} else {
		return out, nil
	}
}
