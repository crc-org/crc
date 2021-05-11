package test_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// =========== Podman ==========

// PodmanBuilder is used to build, customize and execute a podman commands
type PodmanBuilder struct {
	cmd     *exec.Cmd
	timeout <-chan time.Time
}

// NewPodmanCommand returns a PodmanBuilder for running Podman.
func NewPodmanCommand(args ...string) *PodmanBuilder {
	// TODO: split this to `podman-remote` for linux and `podman` for the rest
	cmd := exec.Command("podman", args...)
	return &PodmanBuilder{
		cmd: cmd,
	}
}

// WithTimeout sets the given timeout and returns itself.
func (b *PodmanBuilder) WithTimeout(t <-chan time.Time) *PodmanBuilder {
	b.timeout = t
	return b
}

// ExecWithFullOutput runs the executable and returns the stdout and stderr.
func (b PodmanBuilder) ExecWithFullOutput() (string, string, error) {
	return Exec(b.cmd, b.timeout)
}

// Exec runs the executable.
func (b PodmanBuilder) Exec() (string, error) {
	stdout, _, err := b.ExecWithFullOutput()
	return stdout, err
}

// ExecOrDie runs the executable or dies if error occurs.
func (b PodmanBuilder) ExecOrDie() string {
	stdout, err := b.Exec()
	Expect(err).To(Not(HaveOccurred()))
	return stdout
}

// ExecOrDieWithLogs runs the executable or dies if error occurs.
func (b PodmanBuilder) ExecOrDieWithLogs() (string, string) {
	stdout, stderr, err := b.ExecWithFullOutput()
	Expect(err).To(Not(HaveOccurred()))
	return stdout, stderr
}

//RunPodmanExpectSuccess
func RunPodmanExpectSuccess(args ...string) string {
	return NewPodmanCommand(args...).ExecOrDie()
}

// =========== CRC ==========

// CRCBuilder is used to build, customize and execute a CRC commands
type CRCBuilder struct {
	cmd     *exec.Cmd
	timeout <-chan time.Time
}

// NewCRCCommand returns a CRCBuilder for running CRC.
func NewCRCCommand(args ...string) *CRCBuilder {
	cmd := exec.Command("crc", args...)
	return &CRCBuilder{
		cmd: cmd,
	}
}

// WithEnv sets the given environment and returns itself.
func (b *CRCBuilder) WithEnv(env []string) *CRCBuilder {
	b.cmd.Env = env
	return b
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

// Run command in the shell on the host
func RunOnHost(timeout string, command string, args ...string) (string, string, error) {
	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		fmt.Printf("Failed to parse timeout string: %s, %v", timeout, err)
		return "", "", err
	}
	cmd := exec.Command(command, args...)
	tChan := make(chan time.Time, 1)

	stdout, stderr, err := Exec(cmd, tChan)

	time.Sleep(parsedTimeout)
	tChan <- time.Now()

	return stdout, stderr, err
}

// Run command in the shell on the host (assuming linux bash)
func RunOnHostWithPrivilege(timeout string, args ...string) (string, string, error) {
	parsedTimeout, err := time.ParseDuration(timeout)
	if err != nil {
		fmt.Printf("Failed to parse timeout string: %s, %v", timeout, err)
		return "", "", err
	}
	cmd := exec.Command("sudo", args...)
	tChan := make(chan time.Time, 1)

	stdout, stderr, err := Exec(cmd, tChan)

	time.Sleep(parsedTimeout)
	tChan <- time.Now()

	return stdout, stderr, err
}

// Send command to CRC VM via SSH
func SendCommandToVM(cmd string, vm string, port string) (string, error) {
	pkPath := filepath.Join(os.Getenv("HOME"), ".crc", "machines", "crc", "id_ecdsa")
	pk, _ := ioutil.ReadFile(pkPath)
	signer, err := ssh.ParsePrivateKey(pk)

	if err != nil {
		return "", fmt.Errorf("could not parse the private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User:            "core",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	destination := fmt.Sprintf("%s:%s", vm, port)
	client, err := ssh.Dial("tcp", destination, config)

	if err != nil {
		return "", fmt.Errorf("failed to dial: %v", err)
	}

	// can be multiple of these per 1 client
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create a session: %v", err)
	}
	defer session.Close()

	// Use the session
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("failed to run: %v", err)
	}

	return b.String(), nil
}

func RunCRCDaemon(c chan string) {

	cmd := exec.Command("crc", "daemon")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("error running cmd: %v", err)
	}

	select {
	case status := <-c:
		if status == "done" {
			err = cmd.Process.Kill()
			if err != nil {
				fmt.Printf("error killing the process: %v", err)
			}
		}
	case <-time.After(30 * time.Minute):
		err = cmd.Process.Kill()
		if err != nil {
			fmt.Printf("error killing the process: %v", err)
		}
	}
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
