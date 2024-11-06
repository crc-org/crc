package fakemachine

import (
	"errors"

	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	libmachinehost "github.com/crc-org/crc/v2/pkg/libmachine/host"
	"github.com/crc-org/machine/libmachine/drivers"
)

type FakeSSHClient struct {
	LastExecutedCommand string
	IsSSHClientClosed   bool
}

func (client *FakeSSHClient) Run(command string) ([]byte, []byte, error) {
	client.LastExecutedCommand = command
	return []byte("test"), []byte("test"), nil
}

func (client *FakeSSHClient) Close() {
	client.IsSSHClientClosed = true
}

func NewFakeSSHClient() *FakeSSHClient {
	return &FakeSSHClient{
		LastExecutedCommand: "",
		IsSSHClientClosed:   false,
	}
}

type FakeVirtualMachine struct {
	IsClosed      bool
	IsRemoved     bool
	IsStopped     bool
	FailingStop   bool
	FailingState  bool
	FakeSSHClient *FakeSSHClient
}

func (vm *FakeVirtualMachine) Close() error {
	vm.IsClosed = true
	return nil
}

func (vm *FakeVirtualMachine) Remove() error {
	vm.IsRemoved = true
	return nil
}

func (vm *FakeVirtualMachine) Stop() error {
	if vm.FailingStop {
		return errors.New("stopping failed")
	}
	vm.IsStopped = true
	return nil
}

func (vm *FakeVirtualMachine) State() (state.State, error) {
	if vm.FailingState {
		return state.Error, errors.New("unable to get virtual machine state")
	}
	if vm.IsClosed || vm.IsStopped {
		return state.Stopped, nil
	}
	return state.Running, nil
}

func (vm *FakeVirtualMachine) IP() (string, error) {
	panic("not implemented")
}

func (vm *FakeVirtualMachine) SSHPort() int {
	panic("not implemented")
}

func (vm *FakeVirtualMachine) SSHRunner() (*ssh.Runner, error) {
	runner, err := ssh.CreateRunnerWithClient(vm.FakeSSHClient)
	return runner, err
}

func (vm *FakeVirtualMachine) Bundle() *bundle.CrcBundleInfo {
	panic("not implemented")
}

func (vm *FakeVirtualMachine) Driver() drivers.Driver {
	panic("not implemented")
}

func (vm *FakeVirtualMachine) API() libmachine.API {
	panic("not implemented")
}

func (vm *FakeVirtualMachine) Host() *libmachinehost.Host {
	panic("not implemented")
}

func (vm *FakeVirtualMachine) Kill() error {
	panic("not implemented")
}

func NewFakeVirtualMachine(failingStop bool, failingState bool) *FakeVirtualMachine {
	return &FakeVirtualMachine{
		FakeSSHClient: NewFakeSSHClient(),
		IsStopped:     false,
		IsRemoved:     false,
		IsClosed:      false,
		FailingStop:   failingStop,
		FailingState:  failingState,
	}
}
