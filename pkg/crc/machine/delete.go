package machine

import (
	"github.com/pkg/errors"
)

func (client *client) Delete() error {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return errors.Wrap(err, "Cannot load machine")
	}

	crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return errors.Wrap(err, "Error loading bundle metadata")
	}
	clusterConfig, err := getClusterConfig(crcBundleMetadata)
	if err != nil {
		return errors.Wrap(err, "Cannot get cluster configuration")
	}

	if err := host.Driver.Remove(); err != nil {
		return errors.Wrap(err, "Driver cannot remove machine")
	}

	if err := libMachineAPIClient.Remove(client.name); err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}
	if err := removeContextFromKubeconfig(clusterConfig); err != nil {
		return errors.Wrap(err, "Cannot remove crc cluster entry from kubeconfig")
	}
	return nil
}
