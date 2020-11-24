package systemd

import (
	"errors"
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

func TestSystemdStatuses(t *testing.T) {
	systemctl, _ := newMockCommander(t)

	status, err := systemctl.Status("running.service")
	assert.NoError(t, err)
	assert.Equal(t, states.Running.String(), status.String())

	status, err = systemctl.Status("listening.socket")
	assert.NoError(t, err)
	assert.Equal(t, states.Listening.String(), status.String())

	status, err = systemctl.Status("stopped.service")
	assert.NoError(t, err)
	assert.Equal(t, states.Stopped.String(), status.String())

	status, err = systemctl.Status("notfound.service")
	assert.NoError(t, err)
	assert.Equal(t, states.NotFound.String(), status.String())
}

type mockSystemdRunner struct {
	test    *testing.T
	failing bool
}

func (r *mockSystemdRunner) Run(command string, args ...string) (string, string, error) {
	assertSystemCtlCommand(r.test, "status", command, args)
	assert.GreaterOrEqual(r.test, len(args), 2)

	unitName := args[1]
	switch unitName {
	case "running.service":
		return r.status(states.Running)
	case "listening.socket":
		return r.status(states.Listening)
	case "stopped.service":
		return r.status(states.Stopped)
	case "notfound.service":
		return r.status(states.NotFound)
	default:
		return r.status(states.Running)
	}
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

const (
	statusRunning string = `● running.service - running service
     Loaded: loaded (/usr/lib/systemd/system/running.service; enabled; vendor preset: enabled)
     Active: active (running) since Tue 2020-11-24 14:52:20 CET; 15s ago
TriggeredBy: ● listening.socket
       Docs: man:running(8)
             https://running.example.com
   Main PID: 224516 (running)
      Tasks: 19 (limit: 32768)
     Memory: 36.9M
        CPU: 414ms
     CGroup: /system.slice/running.service
             ├─ 19588 /usr/sbin/true
             ├─ 19589 /usr/sbin/true
             └─224516 /usr/sbin/running
`
	statusListening string = `● listening.socket - listening socket
     Loaded: loaded (/usr/lib/systemd/system/listening.socket; enabled; vendor preset: disabled)
     Active: active (listening) since Mon 2020-11-16 17:02:29 CET; 1 weeks 0 days ago
   Triggers: ● running.service
     Listen: /run/listening/listening-sock (Stream)
      Tasks: 0 (limit: 38160)
     Memory: 0B
        CPU: 0
     CGroup: /system.slice/listening.socket
`
	statusStopped string = `● stopped.service - stopped service
     Loaded: loaded (/usr/lib/systemd/system/stopped.service; enabled; vendor preset: enabled)
     Active: inactive (dead) since Tue 2020-11-17 12:22:15 CET; 1 weeks 0 days ago
TriggeredBy: ● listening.socket
       Docs: man:stopped(8)
             https://stopped.example.com
    Process: 64922 ExecStart=/usr/sbin/true (code=exited, status=0/SUCCESS)
   Main PID: 64922 (code=exited, status=0/SUCCESS)
        CPU: 327ms
`
	statusNotFound string = "Unit notfound.service could not be found."
)

func (r *mockSystemdRunner) status(s states.State) (string, string, error) {
	var (
		err    error
		stdout string
		stderr string
	)

	if r.failing {
		return "error", "", fmt.Errorf("Failed to run systemd command")
	}
	switch s {
	case states.Running:
		stdout = statusRunning
	case states.Listening:
		stdout = statusListening
	case states.Stopped:
		stdout = statusStopped
		err = errors.New("exit code: 3 - see EXIT STATUS in man systemctl")
	case states.NotFound:
		stderr = statusNotFound
		err = errors.New("exit code: 4 - see EXIT STATUS in man systemctl")
	}

	return stdout, stderr, err
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
