package machine

import "fmt"

func (client *client) Exists() (bool, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return false, err
	}
	exists, err := libMachineAPIClient.Exists(client.name)
	if err != nil {
		return false, fmt.Errorf("Error checking if the host exists: %s", err)
	}
	return exists, nil
}
