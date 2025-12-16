package machine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"time"

	gvproxyclient "github.com/containers/gvisor-tap-vsock/pkg/client"
	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/pkg/errors"
)

func exposePorts(gvClient *gvproxyclient.Client, preset crcPreset.Preset, ingressHTTPPort, ingressHTTPSPort uint) error {
	portsToExpose := vsockPorts(preset, ingressHTTPPort, ingressHTTPSPort)
	for _, port := range portsToExpose {
		if err := gvClient.Expose(&port); err != nil {
			return errors.Wrapf(err, "failed to expose port %s -> %s", port.Local, port.Remote)
		}
	}
	return nil
}

func getGVProxyClient(vmName string) (*gvproxyclient.Client, error) {
	m := getMacadamClient()
	socketPath, err := m.GetGVProxySocketPath(vmName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get gvproxy socket path")
	}
	if _, err := os.Stat(socketPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("gvproxy socket does not exist at %s", socketPath)
		}
		return nil, errors.Wrapf(err, "failed to check gvproxy socket at %s", socketPath)
	}

	baseTransport := httpproxy.HTTPTransport()

	var transport *http.Transport
	if t, ok := baseTransport.(*http.Transport); ok {
		transport = t.Clone()
	} else {
		transport = &http.Transport{}
	}

	transport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return net.Dial("unix", socketPath)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	return gvproxyclient.New(client, "http://gvproxy"), nil
}

// enableInternalDNS configures the internal DNS server via gvproxy Unix socket
func enableInternalDNS(gvClient *gvproxyclient.Client) error {
	zone1 := types.Zone{
		Name: "crc.testing.",
		Records: []types.Record{
			{Name: "host", IP: net.ParseIP("192.168.127.254")},
			{Name: "api", IP: net.ParseIP("192.168.127.2")},
			{Name: "api-int", IP: net.ParseIP("192.168.127.2")},
			{Name: "crc", IP: net.ParseIP("192.168.126.11")},
		},
	}

	if err := gvClient.AddDNS(&zone1); err != nil {
		return errors.Wrap(err, "failed to add DNS zone to gvproxy")
	}

	zone2 := types.Zone{
		Name:      "apps-crc.testing.",
		DefaultIP: net.ParseIP("192.168.127.2"),
	}

	if err := gvClient.AddDNS(&zone2); err != nil {
		return errors.Wrap(err, "failed to add DNS zone to gvproxy")
	}

	logging.Info("Successfully configured internal DNS server")
	return nil
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
