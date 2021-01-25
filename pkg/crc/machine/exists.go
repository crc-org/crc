package machine

import "fmt"

func (client *client) Exists() (bool, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
	exists, err := libMachineAPIClient.Exists(client.name)
	if err != nil {
		return false, fmt.Errorf("Error checking if the host exists: %s", err)
	}
	return exists, nil
}
