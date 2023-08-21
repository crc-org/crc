package machine

import (
	"fmt"

	crcErr "github.com/crc-org/crc/v2/pkg/crc/errors"
)

func (client *client) Exists() (bool, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
	exists, err := libMachineAPIClient.Exists(client.name)
	if err != nil {
		return false, fmt.Errorf("Error checking if the host exists: %s", err)
	}
	return exists, nil
}

func CheckIfMachineMissing(client Client) error {
	exists, err := client.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return crcErr.VMNotExist
	}
	return nil
}
