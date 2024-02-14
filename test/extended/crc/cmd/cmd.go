package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/test/extended/util"
	"github.com/sirupsen/logrus"
)

const (
	// timeout to wait for cluster to change its state
	clusterStateRetryCount = 25
	clusterStateTimeout    = 600
	// defines the number of times the state should be matches in a row
	clusterStateRepetition    = 3
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
	return util.MatchRepetitionsWithRetry(state, CheckCRCStatus, clusterStateRepetition,
		clusterStateRetryCount, clusterStateTimeout)
}

func CheckCRCStatus(state string) error {

	// Podman status does not show Running, so need to OR it separately
	expression := `.*CRC VM: *Running\s.*: *Running|.*CRC VM: *Running\s.*Podman: *\d+\.\d+\.\d+`

	// Does not apply to Podman preset
	if state == "stopped" {
		expression = `.*CRC VM:.*\s.*: .*Stopped.*`
	}

	err := util.ExecuteCommand("crc status")
	if err != nil {
		return err
	}
	out := util.GetLastCommandOutput("stdout")
	err = util.CompareExpectedWithActualMatchesRegex(expression, string(out))
	return err
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

// PODMAN

// PodmanBuilder is used to build, customize, and execute a podman-remote command.
type PodmanBuilder struct {
	cmd     *exec.Cmd
	timeout <-chan time.Time
}

// NewPodmanCommand returns a PodmanBuilder for running CRC.
func NewPodmanCommand(args ...string) *PodmanBuilder {

	cmd := exec.Command("podman", args...)

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("podman-remote", args...)
	case "windows":
		cmd = exec.Command("podman.exe", args...)
	}

	return &PodmanBuilder{
		cmd: cmd,
	}
}

// WithTimeout sets the given timeout and returns itself.
func (b *PodmanBuilder) WithTimeout(t <-chan time.Time) *PodmanBuilder {
	b.timeout = t
	return b
}

// WithStdinData sets the given data to stdin and returns itself.
func (b PodmanBuilder) WithStdinData(data string) *PodmanBuilder {
	b.cmd.Stdin = strings.NewReader(data)
	return &b
}

// WithStdinReader sets the given reader and returns itself.
func (b PodmanBuilder) WithStdinReader(reader io.Reader) *PodmanBuilder {
	b.cmd.Stdin = reader
	return &b
}

// ExecOrDie runs the executable or dies if error occurs.
func (b PodmanBuilder) ExecOrDie() (string, error) {
	stdout, err := b.Exec()
	return stdout, err
}

// ExecOrDieWithLogs runs the executable or dies if error occurs.
func (b PodmanBuilder) ExecOrDieWithLogs() (string, string, error) {
	stdout, stderr, err := b.ExecWithFullOutput()
	return stdout, stderr, err
}

// Exec runs the executable.
func (b PodmanBuilder) Exec() (string, error) {
	stdout, _, err := b.ExecWithFullOutput()
	return stdout, err
}

// ExecWithFullOutput runs the executable and returns the stdout and stderr.
func (b PodmanBuilder) ExecWithFullOutput() (string, string, error) {
	return Exec(b.cmd, b.timeout)
}

func Exec(cmd *exec.Cmd, timeout <-chan time.Time) (string, string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	logrus.Infof("Running '%s %s'", cmd.Path, strings.Join(cmd.Args[1:], " ")) // skip arg[0] as it is printed separately
	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("error starting %v:\nCommand stdout:\n%v\nstderr:\n%v\nerror:\n%v", cmd, cmd.Stdout, cmd.Stderr, err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()
	select {
	case err := <-errCh:
		if err != nil {
			var rc = 127
			if ee, ok := err.(*exec.ExitError); ok {
				rc = int(ee.Sys().(syscall.WaitStatus).ExitStatus())
				logrus.Infof("rc: %d", rc)
			}
			return stdout.String(), stderr.String(), CodeExitError{
				Err:  fmt.Errorf("error running %v:\nCommand stdout:\n%v\nstderr:\n%v\nerror:\n%v", cmd, cmd.Stdout, cmd.Stderr, err),
				Code: rc,
			}
		}
	case <-timeout:
		_ = cmd.Process.Kill()
		return "", "", fmt.Errorf("timed out waiting for command %v:\nCommand stdout:\n%v\nstderr:\n%v", cmd, cmd.Stdout, cmd.Stderr)
	}
	logrus.Infof("stderr: %q", stderr.String())
	logrus.Infof("stdout: %q", stdout.String())
	return stdout.String(), stderr.String(), nil
}

// RunPodmanExpectSuccess is a convenience wrapper over podman-remote
func RunPodmanExpectSuccess(args ...string) (string, error) {
	return NewPodmanCommand(args...).ExecOrDie()
}

// RunPodmanExpectFail is a convenience wrapper over PodmanBuilder
// if err != nil: return stderr, nil
// if err == nil: return stdout, err
func RunPodmanExpectFail(args ...string) (string, error) {
	stdout, stderr, err := NewPodmanCommand(args...).ExecWithFullOutput()

	if err == nil {
		err = fmt.Errorf("Expected error but exited without error")
		return stdout, err
	}

	return stderr, nil
}

// ExitError is an interface that presents an API similar to os.ProcessState, which is
// what ExitError from os/exec is.  This is designed to make testing a bit easier and
// probably loses some of the cross-platform properties of the underlying library.
type ExitError interface {
	String() string
	Error() string
	Exited() bool
	ExitStatus() int
}

// CodeExitError is an implementation of ExitError consisting of an error object
// and an exit code (the upper bits of os.exec.ExitStatus).
type CodeExitError struct {
	Err  error
	Code int
}

var _ ExitError = CodeExitError{}

func (e CodeExitError) Error() string {
	return e.Err.Error()
}

func (e CodeExitError) String() string {
	return e.Err.Error()
}

func (e CodeExitError) Exited() bool {
	return true
}

func (e CodeExitError) ExitStatus() int {
	return e.Code
}
