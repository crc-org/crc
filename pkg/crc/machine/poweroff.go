package machine

import "github.com/pkg/errors"

func (client *client) PowerOff() error {
	m := getMacadamClient()
	_, _, err := m.StopVM(client.name)
	if err != nil {
		return errors.Wrap(err, "Cannot kill machine")
	}
	return nil
}
