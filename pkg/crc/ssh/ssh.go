package ssh

import (
	"github.com/code-ready/crc/pkg/crc/constants"
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
	return drivers.RunSSHCommandFromDriver(runner.driver, runner.privateSSHKey, command)
}

func (runner *SSHRunner) SetPrivateKeyPath(path string) {
	runner.privateSSHKey = path
}
