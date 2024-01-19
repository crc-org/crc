package machine

import (
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strconv"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	crcErrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/pkg/errors"
)

func exposePorts(preset crcPreset.Preset, ingressHTTPPort, ingressHTTPSPort uint) error {
	portsToExpose := vsockPorts(preset, ingressHTTPPort, ingressHTTPSPort)
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
		if err := daemonClient.NetworkClient.Unexpose(&types.UnexposeRequest{Protocol: port.Protocol, Local: port.Local}); err != nil {
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
	hostVirtualIP    = "192.168.127.254"
	internalSSHPort  = "22"
	remoteHTTPPort   = "80"
	remoteHTTPSPort  = "443"
	apiPort          = "6443"
	cockpitPort      = "9090"
)

func vsockPorts(preset crcPreset.Preset, ingressHTTPPort, ingressHTTPSPort uint) []types.ExposeRequest {
	socketProtocol := types.UNIX
	socketLocal := constants.GetHostDockerSocketPath()
	if runtime.GOOS == "windows" {
		socketProtocol = types.NPIPE
		socketLocal = constants.DefaultPodmanNamedPipe
	}
	exposeRequest := []types.ExposeRequest{
		{
			Protocol: "tcp",
			Local:    net.JoinHostPort(constants.LocalIP, strconv.Itoa(constants.VsockSSHPort)),
			Remote:   net.JoinHostPort(virtualMachineIP, internalSSHPort),
		},
		{
			Protocol: socketProtocol,
			Local:    socketLocal,
			Remote:   getSSHTunnelURI(),
		},
	}

	switch preset {
	case crcPreset.OpenShift, crcPreset.OKD, crcPreset.Microshift:
		exposeRequest = append(exposeRequest,
			types.ExposeRequest{
				Protocol: "tcp",
				Local:    net.JoinHostPort(constants.LocalIP, apiPort),
				Remote:   net.JoinHostPort(virtualMachineIP, apiPort),
			},
			types.ExposeRequest{
				Protocol: "tcp",
				Local:    fmt.Sprintf(":%d", ingressHTTPSPort),
				Remote:   net.JoinHostPort(virtualMachineIP, remoteHTTPSPort),
			},
			types.ExposeRequest{
				Protocol: "tcp",
				Local:    fmt.Sprintf(":%d", ingressHTTPPort),
				Remote:   net.JoinHostPort(virtualMachineIP, remoteHTTPPort),
			})
	case crcPreset.Podman:
		exposeRequest = append(exposeRequest,
			types.ExposeRequest{
				Protocol: "tcp",
				Local:    net.JoinHostPort(constants.LocalIP, cockpitPort),
				Remote:   net.JoinHostPort(virtualMachineIP, cockpitPort),
			})
	default:
		logging.Errorf("Invalid preset: %s", preset)
	}

	return exposeRequest
}

func getSSHTunnelURI() string {
	u := url.URL{
		Scheme:     "ssh-tunnel",
		User:       url.User("core"),
		Host:       net.JoinHostPort(virtualMachineIP, internalSSHPort),
		Path:       "/run/podman/podman.sock",
		ForceQuery: false,
		RawQuery:   fmt.Sprintf("key=%s", url.QueryEscape(constants.GetPrivateKeyPath())),
	}
	return u.String()
}
