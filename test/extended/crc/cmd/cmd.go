package cmd

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	clicumber "github.com/code-ready/clicumber/testsuite"
	"github.com/code-ready/crc/test/extended/util"
)

const (
	// timeout to wait for cluster to change its state
	clusterStateTimeout       = "900"
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
	command            string
	disableUpdateCheck bool
	disableNTP         bool
}

func CRC(command string) Command {
	return Command{command: command}
}

func (c Command) WithDisableUpdateCheck() Command {
	c.disableUpdateCheck = true
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
		return clicumber.ExecuteCommandSucceedsOrFails(c.ToString(), expectedExit)
	}
	return fmt.Errorf("%s is a valid expected exit status", expectedExit)
}

func (c Command) Execute() error {
	if err := c.validate(); err != nil {
		return err
	}
	return clicumber.ExecuteCommand(c.ToString())
}

func (c Command) env() []string {
	var env []string
	if c.disableUpdateCheck {
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
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func UnsetConfigPropertySucceedsOrFails(property string, expected string) error {
	cmd := "crc config unset " + property
	return clicumber.ExecuteCommandSucceedsOrFails(cmd, expected)
}

func WaitForClusterInState(state string) error {
	retryCount := 15
	iterationDuration, extraDuration, err :=
		util.GetRetryParametersFromTimeoutInSeconds(retryCount, clusterStateTimeout)
	if err != nil {
		return err
	}
	for i := 0; i < retryCount; i++ {
		err := CheckCRCStatus(state)
		if err == nil {
			return nil
		}
		time.Sleep(iterationDuration)
	}
	if extraDuration != 0 {
		time.Sleep(extraDuration)
	}
	return fmt.Errorf("the did not reach the %s state", state)
}

func CheckCRCStatus(state string) error {
	expression := `.*OpenShift: .*Running \(v\d+\.\d+\.\d+.*\).*`
	if state == "stopped" {
		expression = ".*OpenShift: .*Stopped.*"
	}

	err := clicumber.ExecuteCommand(CRC("status").ToString())
	if err != nil {
		return err
	}
	return clicumber.CommandReturnShouldMatch("stdout", expression)
}

func CheckCRCExecutableState(state string) error {
	command := "which crc"
	switch state {
	case CRCExecutableInstalled:
		return clicumber.ExecuteCommandSucceedsOrFails(command, "succeeds")
	case CRCExecutableNotInstalled:
		return clicumber.ExecuteCommandSucceedsOrFails(command, "fails")
	default:
		return fmt.Errorf("%s state is not defined as valid crc executable state", state)
	}
}

func CheckMachineNotExists() error {
	expression := `.*Machine does not exist.*`
	err := clicumber.ExecuteCommand(CRC("status").ToString())
	if err != nil {
		return err
	}
	return clicumber.CommandReturnShouldMatch("stderr", expression)
}

func DeleteCRC() error {

	_ = clicumber.ExecuteCommand(CRC("delete").ToString())

	fmt.Printf("Deleted CRC instance (if one existed).\n")
	return nil
}
