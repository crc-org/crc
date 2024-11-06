package machine

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	libmachinehost "github.com/crc-org/crc/v2/pkg/libmachine/host"
	"github.com/crc-org/machine/libmachine/drivers"
	"github.com/pkg/errors"
)

type VirtualMachine interface {
	Close() error
	Remove() error
	State() (state.State, error)
	IP() (string, error)
	SSHPort() int
	SSHRunner() (*ssh.Runner, error)
	Kill() error
	Stop() error
	Bundle() *bundle.CrcBundleInfo
	Driver() drivers.Driver
	API() libmachine.API
	Host() *libmachinehost.Host
}

type virtualMachine struct {
	name   string
	host   *libmachinehost.Host
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

func loadVirtualMachineLazily(vm VirtualMachine, name string, useVSock bool) (VirtualMachine, error) {
	if vm != nil {
		return vm, nil
	}
	return loadVirtualMachine(name, useVSock)
}

func loadVirtualMachine(name string, useVSock bool) (VirtualMachine, error) {
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
		logging.Debugf("Failed to get bundle metadata: %v", err)
		err = errInvalidBundleMetadata
	}

	return &virtualMachine{
		name:   name,
		host:   libmachineHost,
		bundle: crcBundleMetadata,
		api:    apiClient,
		vsock:  useVSock,
	}, err
}

func (vm *virtualMachine) Close() error {
	return vm.api.Close()
}

func (vm *virtualMachine) Remove() error {
	if err := vm.Driver().Remove(); err != nil {
		return errors.Wrap(err, "Driver cannot remove machine")
	}

	if err := vm.api.Remove(vm.name); err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}

	return nil
}

func (vm *virtualMachine) State() (state.State, error) {
	vmStatus, err := vm.Driver().GetState()
	if err != nil {
		return state.Error, err
	}
	return state.FromMachine(vmStatus), nil
}

func (vm *virtualMachine) IP() (string, error) {
	if vm.vsock {
		return "127.0.0.1", nil
	}
	return vm.Driver().GetIP()
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
	return ssh.CreateRunner(ip, vm.SSHPort(), constants.GetPrivateKeyPath(), constants.GetECDSAPrivateKeyPath(), vm.bundle.GetSSHKeyPath())
}

func (vm *virtualMachine) Bundle() *bundle.CrcBundleInfo {
	return vm.bundle
}

func (vm *virtualMachine) Driver() drivers.Driver {
	return vm.host.Driver
}

func (vm *virtualMachine) API() libmachine.API {
	return vm.api
}

func (vm *virtualMachine) Host() *libmachinehost.Host {
	return vm.host
}

func (vm *virtualMachine) Kill() error {
	return vm.host.Kill()
}

func (vm *virtualMachine) Stop() error {
	return vm.host.Stop()
}
