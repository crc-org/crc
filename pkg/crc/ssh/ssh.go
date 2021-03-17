package ssh

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
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
	commandline := fmt.Sprintf("sudo %s", strings.Join(cmdAndArgs, " "))
	return runner.runSSHCommand(commandline, false)
}

func (runner *Runner) CopyData(data []byte, destFilename string, mode os.FileMode) error {
	logging.Debugf("Creating %s with permissions 0%o in the CRC VM", destFilename, mode)
	base64Data := base64.StdEncoding.EncodeToString(data)
	command := fmt.Sprintf("sudo install -m 0%o /dev/null %s && cat <<EOF | base64 --decode | sudo tee %s\n%s\nEOF", mode, destFilename, destFilename, base64Data)
	_, _, err := runner.RunPrivate(command)

	return err
}

func (runner *Runner) CopyFile(srcFilename string, destFilename string, mode os.FileMode) error {
	data, err := ioutil.ReadFile(srcFilename)
	if err != nil {
		return err
	}
	return runner.CopyData(data, destFilename, mode)
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
err     : %w\n`, command, err)
	}

	return string(stdout), string(stderr), nil
}
