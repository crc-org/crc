package machine

import "github.com/pkg/errors"

func (client *client) IP() (string, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return "", errors.Wrap(err, "Cannot initialize libmachine")
	}
	host, err := libMachineAPIClient.Load(client.name)

	if err != nil {
		return "", errors.Wrap(err, "Cannot load machine")
	}
	ip, err := getIP(host, client.useVSock())
	if err != nil {
		return "", errors.Wrap(err, "Cannot get IP")
	}
	return ip, nil
}
