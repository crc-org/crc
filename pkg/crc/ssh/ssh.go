package ssh

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
)

type Runner struct {
	client Client
}

func CreateRunner(ip string, port int, privateKeys ...string) (*Runner, error) {
	client, err := NewClient(constants.DefaultSSHUser, ip, port, privateKeys...)
	if err != nil {
		return nil, err
	}
	return &Runner{
		client: client,
	}, nil
}

func (runner *Runner) Close() {
	runner.client.Close()
}

func (runner *Runner) Run(cmd string, args ...string) (string, string, error) {
	if len(args) != 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	return runner.runSSHCommand(cmd, false)
}

func (runner *Runner) RunPrivate(cmd string, args ...string) (string, string, error) {
	if len(args) != 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	return runner.runSSHCommand(cmd, true)
}

func (runner *Runner) RunPrivileged(reason string, cmdAndArgs ...string) (string, string, error) {
	logging.Debugf("Using root access: %s", reason)
	commandline := fmt.Sprintf("sudo %s", strings.Join(cmdAndArgs, " "))
	return runner.runSSHCommand(commandline, false)
}

func (runner *Runner) copyDataFull(data []byte, destFilename string, mode os.FileMode, privileged bool) error {
	var sudo string
	if privileged {
		sudo = "sudo "
	}
	logging.Debugf("Creating %s with permissions 0%o in the CRC VM", destFilename, mode)
	base64Data := base64.StdEncoding.EncodeToString(data)
	command := fmt.Sprintf("%sinstall -m 0%o /dev/null %s && cat <<EOF | base64 --decode | %stee %s\n%s\nEOF", sudo, mode, destFilename, sudo, destFilename, base64Data)
	_, _, err := runner.RunPrivate(command)

	return err
}

func (runner *Runner) CopyDataPrivileged(data []byte, destFilename string, mode os.FileMode) error {
	return runner.copyDataFull(data, destFilename, mode, true)
}

func (runner *Runner) CopyData(data []byte, destFilename string, mode os.FileMode) error {
	return runner.copyDataFull(data, destFilename, mode, false)
}

func (runner *Runner) CopyFileFromVM(srcFilename string, destFilename string, mode os.FileMode) error {
	command := fmt.Sprintf("sudo base64 %s", srcFilename)
	stdout, stderr, err := runner.RunPrivate(command)
	if err != nil {
		return fmt.Errorf("Failed to get file content %s : %w", stderr, err)
	}
	rawDecodedText, err := base64.StdEncoding.DecodeString(stdout)
	if err != nil {
		return err
	}
	return os.WriteFile(destFilename, rawDecodedText, mode)
}

func (runner *Runner) CopyFile(srcFilename string, destFilename string, mode os.FileMode) error {
	data, err := os.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	return runner.CopyDataPrivileged(data, destFilename, mode)
}

func (runner *Runner) runSSHCommand(command string, runPrivate bool) (string, string, error) {
	if runPrivate {
		logging.Debugf("Running SSH command: <hidden>")
	} else {
		logging.Debugf("Running SSH command: %s", command)
	}

	stdout, stderr, err := runner.client.Run(command)
	if runPrivate {
		if err != nil {
			logging.Debugf("SSH command failed")
		} else {
			logging.Debugf("SSH command succeeded")
		}
	} else {
		logging.Debugf("SSH command results: err: %v, output: %s", err, string(stdout))
	}

	if err != nil {
		return string(stdout), string(stderr), fmt.Errorf(`ssh command error:
command : %s
err     : %w`+"\n", command, err)
	}

	return string(stdout), string(stderr), nil
}

func (runner *Runner) WaitForConnectivity(ctx context.Context, timeout time.Duration) error {
	checkSSHConnectivity := func() error {
		_, _, err := runner.Run("exit 0")
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	return errors.Retry(ctx, timeout, checkSSHConnectivity, time.Second)
}
