package machine

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	libmachinehost "github.com/crc-org/crc/v2/pkg/libmachine/host"
	"github.com/pkg/errors"
)

type virtualMachine struct {
	name string
	*libmachinehost.Host
	bundle *bundle.CrcBundleInfo
	api    libmachine.API
	vsock  bool
}

type MissingHostError struct {
	name string
}

func errMissingHost(name string) *MissingHostError {
	return &MissingHostError{name: name}
}

func (err *MissingHostError) Error() string {
	return fmt.Sprintf("no such libmachine vm: %s", err.name)
}

var errInvalidBundleMetadata = errors.New("Error loading bundle metadata")

func loadVirtualMachine(name string, useVSock bool) (*virtualMachine, error) {
	apiClient := libmachine.NewClient(constants.MachineBaseDir)
	exists, err := apiClient.Exists(name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot check if machine exists")
	}
	if !exists {
		return nil, errMissingHost(name)
	}

	libmachineHost, err := apiClient.Load(name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}

	crcBundleMetadata, err := getBundleMetadataFromDriver(libmachineHost.Driver)
	if err != nil {
		err = errInvalidBundleMetadata
	}

	return &virtualMachine{
		name:   name,
		Host:   libmachineHost,
		bundle: crcBundleMetadata,
		api:    apiClient,
		vsock:  useVSock,
	}, err
}

func (vm *virtualMachine) Close() error {
	return vm.api.Close()
}

func (vm *virtualMachine) Remove() error {
	if err := vm.Driver.Remove(); err != nil {
		return errors.Wrap(err, "Driver cannot remove machine")
	}

	if err := vm.api.Remove(vm.name); err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}

	return nil
}

func (vm *virtualMachine) State() (state.State, error) {
	vmStatus, err := vm.Driver.GetState()
	if err != nil {
		return state.Error, err
	}
	return state.FromMachine(vmStatus), nil
}

func (vm *virtualMachine) IP() (string, error) {
	if vm.vsock {
		return "127.0.0.1", nil
	}
	return vm.Driver.GetIP()
}

func (vm *virtualMachine) SSHPort() int {
	if vm.vsock {
		return constants.VsockSSHPort
	}
	return constants.DefaultSSHPort
}

func (vm *virtualMachine) SSHRunner() (*ssh.Runner, error) {
	ip, err := vm.IP()
	if err != nil {
		return nil, err
	}
	return ssh.CreateRunner(ip, vm.SSHPort(), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath(), vm.bundle.GetSSHKeyPath())
}
