package systemd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd/actions"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
	crcos "github.com/code-ready/crc/pkg/os"
)

type Commander struct {
	commandRunner crcos.CommandRunner
}

func NewInstanceSystemdCommander(sshRunner *ssh.Runner) *Commander {
	return &Commander{
		commandRunner: ssh.NewRemoteCommandRunner(sshRunner),
	}
}

func NewHostSystemdCommander() *Commander {
	return &Commander{
		commandRunner: crcos.NewLocalCommandRunner(),
	}
}

func (c Commander) Enable(name string) (bool, error) {
	_, err := c.service(name, actions.Enable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c Commander) Disable(name string) (bool, error) {
	_, err := c.service(name, actions.Disable)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c Commander) Reload(name string) (bool, error) {
	_, _ = c.DaemonReload()
	_, err := c.service(name, actions.Reload)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (c Commander) Restart(name string) (bool, error) {
	_, _ = c.DaemonReload()
	_, err := c.service(name, actions.Restart)

	if err != nil {
		return false, err
	}
	return true, nil
}

func (c Commander) Start(name string) (bool, error) {
	_, _ = c.DaemonReload()
	_, err := c.service(name, actions.Start)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c Commander) Stop(name string) (bool, error) {
	_, err := c.service(name, actions.Stop)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c Commander) Status(name string) (states.State, error) {
	return c.service(name, actions.Status)

}

func (c Commander) DaemonReload() (bool, error) {
	stdOut, stdErr, err := c.commandRunner.RunPrivileged("execute systemctl daemon-reload command", "systemctl", "daemon-reload")
	if err != nil {
		return false, fmt.Errorf("Executing systemctl daemon-reload failed: %s %v: %s", stdOut, err, stdErr)
	}
	return true, nil
}

func (c Commander) service(name string, action actions.Action) (states.State, error) {
	stdOut, stdErr, err := c.commandRunner.RunPrivileged("execute systemctl command", "systemctl", action.String(), name)
	if err != nil {
		return states.Error, fmt.Errorf("Executing systemctl action failed: %s %v: %s", stdOut, err, stdErr)
	}

	return states.Compare(stdOut), nil
}
