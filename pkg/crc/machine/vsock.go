package machine

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/gvisor-tap-vsock/pkg/types"
	"github.com/pkg/errors"
)

func exposePorts(portsToExpose []types.ExposeRequest) error {
	daemonClient := daemonclient.New()
	alreadyOpenedPorts, err := daemonClient.NetworkClient.List()
	if err != nil {
		logging.Error("Is 'crc daemon' running? Network mode 'vsock' requires 'crc daemon' to be running, run it manually on different terminal/tab")
		return err
	}
	var missingPorts []types.ExposeRequest
	for _, port := range portsToExpose {
		if !isOpened(alreadyOpenedPorts, port) {
			missingPorts = append(missingPorts, port)
		}
	}
	for i := range missingPorts {
		port := &missingPorts[i]
		if err := daemonClient.NetworkClient.Expose(port); err != nil {
			return errors.Wrapf(err, "failed to expose port %s -> %s", port.Local, port.Remote)
		}
	}
	return nil
}

func isOpened(exposed []types.ExposeRequest, port types.ExposeRequest) bool {
	for _, alreadyOpenedPort := range exposed {
		if port == alreadyOpenedPort {
			return true
		}
	}
	return false
}

const (
	virtualMachineIP = "192.168.127.2"
	internalSSHPort  = 22
	httpPort         = 80
	httpsPort        = 443
	apiPort          = 6443
)

func openshiftPorts() []types.ExposeRequest {
	return append(basePorts(), []types.ExposeRequest{
		{
			Local:  fmt.Sprintf(":%d", apiPort),
			Remote: fmt.Sprintf("%s:%d", virtualMachineIP, apiPort),
		},
		{
			Local:  fmt.Sprintf(":%d", httpsPort),
			Remote: fmt.Sprintf("%s:%d", virtualMachineIP, httpsPort),
		},
		{
			Local:  fmt.Sprintf(":%d", httpPort),
			Remote: fmt.Sprintf("%s:%d", virtualMachineIP, httpPort),
		},
	}...)
}

func basePorts() []types.ExposeRequest {
	return []types.ExposeRequest{
		{
			Local:  fmt.Sprintf(":%d", constants.VsockSSHPort),
			Remote: fmt.Sprintf("%s:%d", virtualMachineIP, internalSSHPort),
		},
	}
}
