package systemd

import (
	"fmt"

	crcos "github.com/code-ready/crc/pkg/os"

	"github.com/code-ready/crc/pkg/crc/systemd/actions"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
)

type SystemdCommander interface {
	Start(name string) (bool, error)
	Stop(name string) (bool, error)
	Reload(name string) (bool, error)
	Restart(name string) (bool, error)
	Status(name string) (states.State, error)
	Enable(name string) (bool, error)
	Disable(name string) (bool, error)
	DaemonReload() (bool, error)
}

type HostSystemdCommander struct {
	commander SystemdCommander
}

func NewHostSystemdCommander() *HostSystemdCommander {
	return &HostSystemdCommander{}
}

func (c HostSystemdCommander) Enable(name string) (bool, error) {
	_, err := c.service(name, actions.Enable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c HostSystemdCommander) Disable(name string) (bool, error) {
	_, err := c.service(name, actions.Disable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c HostSystemdCommander) DaemonReload() (bool, error) {
	stdOut, stdErr, err := crcos.RunWithPrivilege("systemctl", "daemon-reload")
	if err != nil {
		return false, fmt.Errorf("Executing systemctl daemon-reload failed: %s %v: %s", stdOut, err, stdErr)
	}
	return true, nil
}

func (c HostSystemdCommander) Reload(name string) (bool, error) {
	c.DaemonReload()
	_, err := c.service(name, actions.Reload)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (c HostSystemdCommander) Restart(name string) (bool, error) {
	c.DaemonReload()
	_, err := c.service(name, actions.Restart)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (c HostSystemdCommander) Start(name string) (bool, error) {
	c.DaemonReload()
	_, err := c.service(name, actions.Start)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c HostSystemdCommander) Stop(name string) (bool, error) {
	_, err := c.service(name, actions.Stop)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c HostSystemdCommander) Status(name string) (states.State, error) {
	return c.service(name, actions.Status)

}

func (c HostSystemdCommander) service(name string, action actions.Action) (states.State, error) {
	stdOut, stdErr, err := crcos.RunWithPrivilege("systemctl", action.String(), name)
	if err != nil {
		return states.Error, fmt.Errorf("Executing systemctl action failed: %s %v: %s", stdOut, err, stdErr)
	}

	return states.Compare(stdOut), nil
}
