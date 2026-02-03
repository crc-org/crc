package systemd

import (
	"fmt"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/crc/systemd/actions"
	"github.com/crc-org/crc/v2/pkg/crc/systemd/results"
	"github.com/crc-org/crc/v2/pkg/crc/systemd/states"
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

type Commander struct {
	commandRunner crcos.CommandRunner
}

func NewInstanceSystemdCommander(sshRunner *ssh.Runner) *Commander {
	return &Commander{
		commandRunner: sshRunner,
	}
}

func (c Commander) Enable(name string) error {
	_, err := c.service(name, actions.Enable)
	return err
}

func (c Commander) Disable(name string) error {
	_, err := c.service(name, actions.Disable)
	return err
}

func (c Commander) Reload(name string) error {
	_ = c.DaemonReload()
	_, err := c.service(name, actions.Reload)
	return err
}

func (c Commander) Restart(name string) error {
	_ = c.DaemonReload()
	_, err := c.service(name, actions.Restart)
	return err
}

func (c Commander) Start(name string) error {
	_ = c.DaemonReload()
	_, err := c.service(name, actions.Start)
	return err
}

func (c Commander) Stop(name string) error {
	_, err := c.service(name, actions.Stop)
	return err
}

func (c Commander) Status(name string) (states.State, error) {
	return c.service(name, actions.Status)

}

// Result returns the result of a service execution (success, exit-code, etc.)
func (c Commander) Result(name string) (results.Result, error) {
	stdOut, stdErr, err := c.commandRunner.Run("systemctl", "show", "--property", "Result", "--value", name)
	if err != nil {
		return results.Unknown, fmt.Errorf("failed to get service result: %s %v: %s", stdOut, err, stdErr)
	}

	// Output format is "Result=success" or "Result=exit-code", etc.
	output := strings.TrimSpace(stdOut)
	return results.Parse(output), nil
}

// ExecMainExitTimestamp returns the exit timestamp of a service's main process.
// Returns empty string if the service has never run, otherwise returns the exit timestamp.
func (c Commander) ExecMainExitTimestamp(name string) (string, error) {
	stdOut, stdErr, err := c.commandRunner.Run("systemctl", "show", "--property", "ExecMainExitTimestamp", "--value", name)
	if err != nil {
		return "", fmt.Errorf("failed to get service main exit timestamp: %s %v: %s", stdOut, err, stdErr)
	}
	return strings.TrimSpace(stdOut), nil
}

// WasSkippedDueToConditions returns true if the service was triggered but skipped
// because its condition checks (e.g., ConditionPathExists) were not met.
// This is determined by checking both ConditionResult and ConditionTimestamp:
// - If ConditionTimestamp is empty, conditions were never evaluated (service never triggered)
// - If ConditionTimestamp has a value and ConditionResult is "no", service was skipped
func (c Commander) WasSkippedDueToConditions(name string) (bool, error) {
	stdOut, stdErr, err := c.commandRunner.Run("systemctl", "show", "-p", "ConditionResult", "-p", "ConditionTimestamp", name)
	if err != nil {
		return false, fmt.Errorf("failed to get service condition info: %s %v: %s", stdOut, err, stdErr)
	}

	output := strings.TrimSpace(stdOut)
	var conditionResult, conditionTimestamp string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "ConditionResult=") {
			conditionResult = strings.TrimPrefix(line, "ConditionResult=")
		} else if strings.HasPrefix(line, "ConditionTimestamp=") {
			conditionTimestamp = strings.TrimPrefix(line, "ConditionTimestamp=")
		}
	}

	// Service was skipped only if conditions were actually evaluated (timestamp exists)
	// AND the result was "no" (conditions not met)
	return conditionTimestamp != "" && conditionResult == "no", nil
}

func (c Commander) DaemonReload() error {
	stdOut, stdErr, err := c.commandRunner.RunPrivileged("Executing systemctl daemon-reload command", "systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("Executing systemctl daemon-reload failed: %s %v: %s", stdOut, err, stdErr)
	}
	return nil
}

func (c Commander) service(name string, action actions.Action) (states.State, error) {
	var (
		stdOut, stdErr string
		err            error
	)

	if action.IsPriviledged() {
		msg := fmt.Sprintf("Executing systemctl %s %s", action.String(), name)
		stdOut, stdErr, err = c.commandRunner.RunPrivileged(msg, "systemctl", action.String(), name)
	} else {
		stdOut, stdErr, err = c.commandRunner.Run("systemctl", action.String(), name)
	}

	if err != nil {
		state := states.Compare(stdOut)
		if state != states.Unknown {
			return state, nil
		}
		state = states.Compare(stdErr)
		if state == states.NotFound {
			return state, nil
		}

		return states.Error, fmt.Errorf("Executing systemctl action failed: %s %v: %s", stdOut, err, stdErr)
	}

	return states.Compare(stdOut), nil
}

type systemctlUserRunner struct {
	runner crcos.CommandRunner
}

func (userRunner *systemctlUserRunner) Run(command string, args ...string) (string, string, error) {
	if command != "systemctl" {
		return "", "", fmt.Errorf("Invalid command: '%s'", command)
	}
	return userRunner.runner.Run("systemctl", append([]string{"--user"}, args...)...)
}

func (userRunner *systemctlUserRunner) RunPrivate(command string, args ...string) (string, string, error) {
	if command != "systemctl" {
		return "", "", fmt.Errorf("Invalid command: '%s'", command)
	}
	return userRunner.runner.RunPrivate("systemctl", append([]string{"--user"}, args...)...)
}

func (userRunner *systemctlUserRunner) RunPrivileged(_ string, cmdAndArgs ...string) (string, string, error) {
	command := cmdAndArgs[0]
	args := cmdAndArgs[1:]
	if command != "systemctl" {
		return "", "", fmt.Errorf("Invalid command: '%s'", command)
	}
	return userRunner.runner.Run("systemctl", append([]string{"--user"}, args...)...)
}

func (c *Commander) User() *Commander {
	return &Commander{
		commandRunner: &systemctlUserRunner{
			c.commandRunner,
		},
	}
}
