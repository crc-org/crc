package ssh

import (
	"fmt"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/machine/libmachine/drivers"
)

type SSHRunner struct {
	driver        drivers.Driver
	privateSSHKey string
}

func CreateRunner(driver drivers.Driver) *SSHRunner {
	return CreateRunnerWithPrivateKey(driver, constants.GetPrivateKeyPath())
}

func CreateRunnerWithPrivateKey(driver drivers.Driver, privateKey string) *SSHRunner {
	return &SSHRunner{driver: driver, privateSSHKey: privateKey}
}

// Create a host using the driver's config
func (runner *SSHRunner) Run(command string) (string, error) {
	return runner.runSSHCommandFromDriver(command, false)
}

func (runner *SSHRunner) SetPrivateKeyPath(path string) {
	runner.privateSSHKey = path
}

func (runner *SSHRunner) runSSHCommandFromDriver(command string, runPrivate bool) (string, error) {
	client, err := drivers.GetSSHClientFromDriver(runner.driver, runner.privateSSHKey)
	if err != nil {
		return "", err
	}

	if runPrivate {
		logging.Debugf("About to run SSH command with hidden output")
	} else {
		logging.Debugf("About to run SSH command:\n%s", command)
	}

	output, err := client.Output(command)
	if !runPrivate {
		logging.Debugf("SSH cmd err, output: %v: %s", err, output)
	}
	if err != nil {
		return "", fmt.Errorf(`ssh command error:
command : %s
err     : %v
output  : %s`, command, err, output)
	}

	return output, nil
}
