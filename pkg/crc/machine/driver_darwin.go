package machine

import (
	"encoding/json"
	"errors"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/hyperkit"
	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/host"
)

func newHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	json, err := json.Marshal(hyperkit.CreateHost(machineConfig))
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}
	return api.NewHost("hyperkit", constants.CrcBinDir, json)
}
