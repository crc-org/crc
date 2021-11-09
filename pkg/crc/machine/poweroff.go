package machine

import "github.com/pkg/errors"

func (client *client) PowerOff() error {
	vm, err := loadVirtualMachine(client.name, client.useVSock())
	if err != nil {
		return errors.Wrap(err, "Cannot load machine")
	}
	defer vm.Close()

	if err := vm.Kill(); err != nil {
		return errors.Wrap(err, "Cannot kill machine")
	}
	return nil
}
