package ssh

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/machine/libmachine/drivers"
)

type Runner struct {
	driver        drivers.Driver
	privateSSHKey string
}

func CreateRunner(driver drivers.Driver) *Runner {
	return CreateRunnerWithPrivateKey(driver, constants.GetPrivateKeyPath())
}

func CreateRunnerWithPrivateKey(driver drivers.Driver, privateKey string) *Runner {
	return &Runner{driver: driver, privateSSHKey: privateKey}
}

// Create a host using the driver's config
func (runner *Runner) Run(command string) (string, error) {
	return runner.runSSHCommandFromDriver(command, false)
}

func (runner *Runner) RunPrivate(command string) (string, error) {
	return runner.runSSHCommandFromDriver(command, true)
}

func (runner *Runner) SetPrivateKeyPath(path string) {
	runner.privateSSHKey = path
}

func (runner *Runner) CopyData(data []byte, destFilename string, mode os.FileMode) error {
	logging.Debugf("Creating %s with permissions 0%o in the CRC VM", destFilename, mode)
	base64Data := base64.StdEncoding.EncodeToString(data)
	command := fmt.Sprintf("sudo install -m 0%o /dev/null %s && cat <<EOF | base64 --decode | sudo tee %s\n%s\nEOF", mode, destFilename, destFilename, base64Data)
	_, err := runner.RunPrivate(command)

	return err
}

func (runner *Runner) runSSHCommandFromDriver(command string, runPrivate bool) (string, error) {
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
	if runPrivate {
		if err != nil {
			logging.Debugf("SSH command failed")
		} else {
			logging.Debugf("SSH command succeeded")
		}
	} else {
		logging.Debugf("SSH command results: err: %v, output: %s", err, output)
	}

	if err != nil {
		return "", fmt.Errorf(`ssh command error:
command : %s
err     : %v
output  : %s`, command, err, output)
	}

	return output, nil
}
