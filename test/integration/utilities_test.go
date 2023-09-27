package test_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/crc-org/crc/v2/test/extended/crc/cmd"
	. "github.com/onsi/gomega"
)

// CRCBuilder is used to build, customize and execute a CRC command.
type CRCBuilder struct {
	cmd     *exec.Cmd
	timeout <-chan time.Time
}

// NewCRCCommand returns a CRCBuilder for running CRC.
func NewCRCCommand(args ...string) *CRCBuilder {
	cmd := exec.Command("crc", args...)
	cmd.Env = append(os.Environ(), "CRC_DISABLE_UPDATE_CHECK=true")
	cmd.Env = append(os.Environ(), "CRC_LOG_LEVEL=debug")
	return &CRCBuilder{
		cmd: cmd,
	}
}

// WithTimeout sets the given timeout and returns itself.
func (b *CRCBuilder) WithTimeout(t <-chan time.Time) *CRCBuilder {
	b.timeout = t
	return b
}

// WithStdinData sets the given data to stdin and returns itself.
func (b CRCBuilder) WithStdinData(data string) *CRCBuilder {
	b.cmd.Stdin = strings.NewReader(data)
	return &b
}

// WithStdinReader sets the given reader and returns itself.
func (b CRCBuilder) WithStdinReader(reader io.Reader) *CRCBuilder {
	b.cmd.Stdin = reader
	return &b
}

// ExecOrDie runs the executable or dies if error occurs.
func (b CRCBuilder) ExecOrDie() string {
	stdout, err := b.Exec()
	Expect(err).To(Not(HaveOccurred()))
	return stdout
}

// ExecOrDieWithLogs runs the executable or dies if error occurs.
func (b CRCBuilder) ExecOrDieWithLogs() (string, string) {
	stdout, stderr, err := b.ExecWithFullOutput()
	Expect(err).To(Not(HaveOccurred()))
	return stdout, stderr
}

// Exec runs the executable.
func (b CRCBuilder) Exec() (string, error) {
	stdout, _, err := b.ExecWithFullOutput()
	return stdout, err
}

// ExecWithFullOutput runs the executable and returns the stdout and stderr.
func (b CRCBuilder) ExecWithFullOutput() (string, string, error) {
	return cmd.Exec(b.cmd, b.timeout)
}

// RunCRCExpectSuccess is a convenience wrapper over CRC
func RunCRCExpectSuccess(args ...string) string {
	return NewCRCCommand(args...).ExecOrDie()
}

// RunCRCExpectFail is a convenience wrapper over CRCBuilder
// if err != nil: return stderr, nil
// if err == nil: return stdout, err
func RunCRCExpectFail(args ...string) (string, error) {
	stdout, stderr, err := NewCRCCommand(args...).ExecWithFullOutput()

	if err == nil {
		err = fmt.Errorf("Expected error but exited without error")
		return stdout, err
	}

	return stderr, nil
}
