package machine

import (
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/pkg/errors"
)

func (client *client) ConnectionDetails() (*types.ConnectionDetails, error) {
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()

	ip, err := vm.IP()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get IP")
	}
	return &types.ConnectionDetails{
		IP:          ip,
		SSHPort:     vm.SSHPort(),
		SSHUsername: constants.DefaultSSHUser,
		SSHKeys:     []string{constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath(), vm.bundle.GetSSHKeyPath()},
	}, nil
}
