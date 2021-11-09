package machine

import (
	"github.com/code-ready/crc/pkg/crc/machine/state"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/pkg/errors"
)

// Return console URL if the VM is present.
func (client *client) GetConsoleURL() (*types.ConsoleResult, error) {
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	vm, err := loadVirtualMachine(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()

	vmState, err := vm.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the state of the virtual machine")
	}

	clusterConfig, err := getClusterConfig(vm.bundle)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading cluster configuration")
	}

	return &types.ConsoleResult{
		ClusterConfig: *clusterConfig,
		State:         state.FromMachine(vmState),
	}, nil
}
