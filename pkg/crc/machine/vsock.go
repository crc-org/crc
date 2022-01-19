package machine

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcPreset "github.com/code-ready/crc/pkg/crc/preset"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/pkg/errors"
)

func exposePorts(preset crcPreset.Preset) error {
	portsToExpose := vsockPorts(preset)
	daemonClient := daemonclient.New()
	alreadyOpenedPorts, err := listOpenPorts(daemonClient)
	if err != nil {
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

func unexposePorts() error {
	var mErr crcErrors.MultiError
	daemonClient := daemonclient.New()
	alreadyOpenedPorts, err := listOpenPorts(daemonClient)
	if err != nil {
		return err
	}
	for _, port := range alreadyOpenedPorts {
		if err := daemonClient.NetworkClient.Unexpose(&types.UnexposeRequest{Local: port.Local}); err != nil {
			mErr.Collect(errors.Wrapf(err, "failed to unexpose port %s ", port.Local))
		}
	}
	if len(mErr.Errors) == 0 {
		return nil
	}
	return mErr
}

func listOpenPorts(daemonClient *daemonclient.Client) ([]types.ExposeRequest, error) {
	alreadyOpenedPorts, err := daemonClient.NetworkClient.List()
	if err != nil {
		logging.Error("Is 'crc daemon' running? Network mode 'vsock' requires 'crc daemon' to be running, run it manually on different terminal/tab")
		return nil, err
	}
	return alreadyOpenedPorts, nil
}

const (
	virtualMachineIP = "192.168.127.2"
	internalSSHPort  = 22
	localIP          = "127.0.0.1"
	httpPort         = 80
	httpsPort        = 443
	apiPort          = 6443
	cockpitPort      = 9090
)

func vsockPorts(preset crcPreset.Preset) []types.ExposeRequest {
	exposeRequest := []types.ExposeRequest{
		{
			Local:  fmt.Sprintf("%s:%d", localIP, constants.VsockSSHPort),
			Remote: fmt.Sprintf("%s:%d", virtualMachineIP, internalSSHPort),
		},
	}
	switch preset {
	case crcPreset.OpenShift:
		exposeRequest = append(exposeRequest,
			types.ExposeRequest{
				Local:  fmt.Sprintf("%s:%d", localIP, apiPort),
				Remote: fmt.Sprintf("%s:%d", virtualMachineIP, apiPort),
			},
			types.ExposeRequest{
				Local:  fmt.Sprintf(":%d", httpsPort),
				Remote: fmt.Sprintf("%s:%d", virtualMachineIP, httpsPort),
			},
			types.ExposeRequest{
				Local:  fmt.Sprintf(":%d", httpPort),
				Remote: fmt.Sprintf("%s:%d", virtualMachineIP, httpPort),
			})
	case crcPreset.Podman:
		exposeRequest = append(exposeRequest,
			types.ExposeRequest{
				Local:  fmt.Sprintf("%s:%d", localIP, cockpitPort),
				Remote: fmt.Sprintf("%s:%d", virtualMachineIP, cockpitPort),
			})
	default:
		logging.Errorf("Invalid preset: %s", preset)
	}

	return exposeRequest
}
