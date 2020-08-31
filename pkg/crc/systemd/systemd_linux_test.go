package systemd

import (
	"fmt"
	"testing"

	"github.com/code-ready/crc/pkg/crc/systemd/states"

	"github.com/stretchr/testify/assert"
)

func newMockCommander(t *testing.T) (*Commander, *mockSystemdRunner) {
	runner := mockSystemdRunner{
		test: t,
	}

	return &Commander{
		commandRunner: &runner,
	}, &runner
}

func TestSystemd(t *testing.T) {
	systemctl, mockRunner := newMockCommander(t)
	err := systemctl.Enable("foo")
	assert.NoError(t, err)

	status, err := systemctl.Status("foo")
	assert.NoError(t, err)
	assert.Equal(t, states.Running, status)

	mockRunner.setFailing(true)

	err = systemctl.Disable("foo")
	assert.Error(t, err)
}

type mockSystemdRunner struct {
	test    *testing.T
	failing bool
}

func (r *mockSystemdRunner) Run(command string, args ...string) (string, string, error) {
	assertSystemCtlCommand(r.test, "status", command, args)
	return r.status(states.Running)
}

func (r *mockSystemdRunner) RunPrivate(command string, args ...string) (string, string, error) {
	r.test.FailNow()
	return "", "", fmt.Errorf("Unexpected RunPrivate() call")
}

func (r *mockSystemdRunner) RunPrivileged(reason string, cmdAndArgs ...string) (string, string, error) {
	privilegedCommands := []string{
		"start",
		"stop",
		"reload",
		"restart",
		"enable",
		"disable",
		"daemon-reload",
	}
	assert.GreaterOrEqual(r.test, len(cmdAndArgs), 2)
	assertSystemCtlCommands(r.test, privilegedCommands, cmdAndArgs[0], cmdAndArgs[1:])

	return r.status(states.Running)
}

func (r *mockSystemdRunner) setFailing(failing bool) {
	r.failing = failing
}

func (r *mockSystemdRunner) status(s states.State) (string, string, error) {
	var (
		err    error
		stdout string
	)

	if r.failing {
		stdout = "error"
		err = fmt.Errorf("Failed to run systemd command")
	} else {
		stdout = s.String()
	}

	return stdout, "", err
}

func assertSystemCtlCommands(t *testing.T, systemctlCommands []string, cmd string, args []string) {
	assert := assert.New(t)

	assert.Equal(cmd, "systemctl")
	assert.Greater(len(args), 0)
	assert.Contains(systemctlCommands, args[0])
}

func assertSystemCtlCommand(t *testing.T, systemctlCommand string, cmd string, args []string) {
	assertSystemCtlCommands(t, []string{systemctlCommand}, cmd, args)
}
