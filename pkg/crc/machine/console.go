package machine

import "github.com/pkg/errors"

// Return console URL if the VM is present.
func (client *client) GetConsoleURL() (*ConsoleResult, error) {
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot initialize libmachine")
	}
	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}

	vmState, err := host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the state for host")
	}

	_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bundle metadata")
	}

	clusterConfig, err := getClusterConfig(crcBundleMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading cluster configuration")
	}

	return &ConsoleResult{
		ClusterConfig: *clusterConfig,
		State:         vmState,
	}, nil
}
