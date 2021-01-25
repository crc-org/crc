package machine

import "github.com/pkg/errors"

func (client *client) IP() (string, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
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
