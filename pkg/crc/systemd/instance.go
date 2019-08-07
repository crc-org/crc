package systemd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd/actions"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
)

type InstanceSystemdCommander struct {
	sshRunner *ssh.Runner
}

// NewVmSystemdCommander creates a new instance of a VmSystemdCommander
func NewInstanceSystemdCommander(sshRunner *ssh.Runner) *InstanceSystemdCommander {
	return &InstanceSystemdCommander{
		sshRunner: sshRunner,
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
	_, err := c.sshRunner.Run("sudo systemctl daemon-reload")
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

func (c InstanceSystemdCommander) Status(name string) (states.State, error) {
	return c.service(name, actions.Status)

}

func (c InstanceSystemdCommander) IsActive(name string) (bool, error) {
	_, err := c.service(name, actions.IsActive)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c InstanceSystemdCommander) service(name string, action actions.Action) (states.State, error) {
	command := fmt.Sprintf("sudo systemctl -f %s %s", action.String(), name)
	out, err := c.sshRunner.Run(command)
	if err != nil {
		return states.Error, err
	}

	return states.Compare(out), nil
}
