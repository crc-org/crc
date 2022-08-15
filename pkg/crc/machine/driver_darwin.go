package machine

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/vfkit"
	machineVf "github.com/code-ready/crc/pkg/drivers/vfkit"
	"github.com/code-ready/crc/pkg/libmachine"
	"github.com/code-ready/crc/pkg/libmachine/host"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(vfkit.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("vf", constants.BinDir(), json)
}

func loadDriverConfig(host *host.Host) (*machineVf.Driver, error) {
	var vfDriver machineVf.Driver
	err := json.Unmarshal(host.RawDriver, &vfDriver)

	return &vfDriver, err
}

func updateDriverConfig(host *host.Host, driver *machineVf.Driver) error {
	driverData, err := json.Marshal(driver)
	if err != nil {
		return err
	}

	return host.UpdateConfig(driverData)
}

func updateKernelArgs(vm *virtualMachine) error {
	logging.Info("Updating kernel args...")
	sshRunner, err := vm.SSHRunner()
	if err != nil {
		return err
	}
	defer sshRunner.Close()

	if err := sshRunner.WaitForConnectivity(context.Background(), 20*time.Second); err != nil {
		return err
	}

	stdout, stderr, err := sshRunner.RunPrivileged("Get kernel args", `-- sh -c 'rpm-ostree kargs'`)
	if err != nil {
		logging.Errorf("Failed to get kernel args: %v - %s", err, stderr)
		return err
	}
	logging.Debugf("Kernel args: %s", stdout)

	vfkitDriver, err := loadDriverConfig(vm.Host)
	if err != nil {
		return err
	}
	logging.Debugf("Current Kernel args: %s", vfkitDriver.Cmdline)
	vfkitDriver.Cmdline = stdout

	if err := updateDriverConfig(vm.Host, vfkitDriver); err != nil {
		return err
	}
	return vm.api.Save(vm.Host)
}
