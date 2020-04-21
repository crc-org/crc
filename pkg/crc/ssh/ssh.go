package ssh

import (
	"encoding/base64"
	"fmt"
	"os"

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

func (runner *SSHRunner) SetTextContentAsRoot(destFilename string, content string, mode os.FileMode) error {
	logging.Debugf("Creating %s with permissions 0%o in the CRC VM", destFilename, mode)
	command := fmt.Sprintf("sudo install -m 0%o /dev/null %s && cat <<EOF | sudo tee %s\n%s\nEOF", mode, destFilename, destFilename, content)
	_, err := runner.RunPrivate(command)

	return err
}

func (runner *SSHRunner) RunPrivate(command string) (string, error) {
	return runner.runSSHCommandFromDriver(command, true)
}

func (runner *SSHRunner) SetPrivateKeyPath(path string) {
	runner.privateSSHKey = path
}

func (runner *SSHRunner) CopyData(data []byte, destFilename string) error {
	base64Data := base64.StdEncoding.EncodeToString(data)
	command := fmt.Sprintf("echo %s | base64 --decode | sudo tee %s > /dev/null", base64Data, destFilename)
	_, err := runner.Run(command)

	return err
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
