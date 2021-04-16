package machine

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/pkg/errors"
)

func (client *client) ConnectionDetails() (*types.ConnectionDetails, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}

	bundle, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bundle metadata")
	}

	ip, err := getIP(host, client.useVSock())
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get IP")
	}
	return &types.ConnectionDetails{
		IP:          ip,
		SSHPort:     getSSHPort(client.useVSock()),
		SSHUsername: constants.DefaultSSHUser,
		SSHKeys:     []string{constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath(), bundle.GetSSHKeyPath()},
	}, nil
}
