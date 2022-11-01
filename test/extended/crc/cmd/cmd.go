package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/crc-org/crc/test/extended/util"
)

const (
	// timeout to wait for cluster to change its state
	clusterStateRetryCount    = 15
	clusterStateTimeout       = 600
	CRCExecutableInstalled    = "installed"
	CRCExecutableNotInstalled = "notInstalled"
)

var (
	commands = map[string]struct{}{
		"bundle":     {},
		"cleanup":    {},
		"config":     {},
		"console":    {},
		"delete":     {},
		"help":       {},
		"ip":         {},
		"oc-env":     {},
		"podman-env": {},
		"setup":      {},
		"start":      {},
		"status":     {},
		"stop":       {},
		"version":    {},
	}
)

type Command struct {
	command     string
	updateCheck bool
	disableNTP  bool
}

func CRC(command string) Command {
	return Command{command: command}
}

func (c Command) WithUpdateCheck() Command {
	c.updateCheck = true
	return c
}

func (c Command) WithDisableNTP() Command {
	c.disableNTP = true
	return c
}

func (c Command) ToString() string {
	cmd := append(c.env(), "crc", c.command)
	return strings.Join(cmd, " ")
}

func (c Command) ExecuteWithExpectedExit(expectedExit string) error {
	if err := c.validate(); err != nil {
		return err
	}
	if expectedExit == "succeeds" || expectedExit == "fails" {
		return util.ExecuteCommandSucceedsOrFails(c.ToString(), expectedExit)
	}
	return fmt.Errorf("%s is a valid expected exit status", expectedExit)
}

func (c Command) Execute() error {
	if err := c.validate(); err != nil {
		return err
	}
	return util.ExecuteCommand(c.ToString())
}

func (c Command) env() []string {
	var env []string
	if !c.updateCheck {
		env = append(env, envVariable("CRC_DISABLE_UPDATE_CHECK", "true"))
	}
	if c.disableNTP {
		env = append(env, envVariable("CRC_DEBUG_ENABLE_STOP_NTP", "true"))
	}
	return env
}

func envVariable(key, value string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("$env:%s=%s;", key, value)
	}
	return fmt.Sprintf("%s=%s", key, value)
}

func (c Command) validate() error {
	cmdline := strings.Fields(c.command)
	if len(cmdline) < 1 {
		return fmt.Errorf("empty command? %s", c.command)
	}
	if _, ok := commands[cmdline[0]]; !ok {
		return fmt.Errorf("%s is not a supported command", cmdline[0])
	}
	return nil
}

func SetConfigPropertyToValueSucceedsOrFails(property string, value string, expected string) error {
	cmd := "crc config set " + property + " " + value
	return util.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func UnsetConfigPropertySucceedsOrFails(property string, expected string) error {
	cmd := "crc config unset " + property
	return util.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func WaitForClusterInState(state string) error {
	return util.MatchWithRetry(state, CheckCRCStatus,
		clusterStateRetryCount, clusterStateTimeout)
}

func CheckCRCStatus(state string) error {
	expression := `.*OpenShift: .*Running \(v\d+\.\d+\.\d+.*\).*`
	if state == "stopped" {
		expression = ".*OpenShift: .*Stopped.*"
	}

	err := util.ExecuteCommand(CRC("status").ToString())
	if err != nil {
		return err
	}
	return util.CommandReturnShouldMatch("stdout", expression)
}

func CheckCRCExecutableState(state string) error {
	command := "which crc"
	if runtime.GOOS == "windows" {
		if err := util.ExecuteCommand("$env:Path = [System.Environment]::GetEnvironmentVariable(\"Path\",\"Machine\")"); err != nil {
			return err
		}
	}
	switch state {
	case CRCExecutableInstalled:
		return util.ExecuteCommandSucceedsOrFails(command, "succeeds")
	case CRCExecutableNotInstalled:
		return util.ExecuteCommandSucceedsOrFails(command, "fails")
	default:
		return fmt.Errorf("%s state is not defined as valid crc executable state", state)
	}
}

func CheckMachineNotExists() error {
	expression := `.*Machine does not exist.*`
	err := util.ExecuteCommand(CRC("status").ToString())
	if err != nil {
		return err
	}
	return util.CommandReturnShouldMatch("stderr", expression)
}

func DeleteCRC() error {

	_ = util.ExecuteCommand(CRC("delete").ToString())

	fmt.Printf("Deleted CRC instance (if one existed).\n")
	return nil
}

func (c Command) ExecuteSingleWithExpectedExit(expectedExit string) error {
	if err := c.validate(); err != nil {
		return err
	}
	if expectedExit == "succeeds" || expectedExit == "fails" {
		// Disable G204 lint check as it will force us to use fixed args for the command
		cmd := exec.Command("crc", strings.Split(c.command, " ")...) // #nosec G204
		err := cmd.Run()
		logging.Debugf("Running single command crc %s", c.command)
		if err != nil && expectedExit == "fails" ||
			err == nil && expectedExit == "succeeds" {
			return nil
		}
		return fmt.Errorf("%s expected %s but it did not", c.ToString(), expectedExit)
	}
	return fmt.Errorf("%s is a valid expected exit status", expectedExit)
}
