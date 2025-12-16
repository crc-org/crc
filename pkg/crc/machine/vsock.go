package machine

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	// Construct the gvproxy socket path
	// Path format: /run/user/<uid>/macadam/qemu/<vmname>-gvproxy-api.sock
	uid := os.Getuid()
	socketPath := filepath.Join("/run", "user", strconv.Itoa(uid), "macadam", "qemu", fmt.Sprintf("%s-gvproxy-api.sock", vmName))

	// Check if the socket path exists
	if _, err := os.Stat(socketPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("gvproxy socket does not exist at %s", socketPath)
		}
		return nil, errors.Wrapf(err, "failed to check gvproxy socket at %s", socketPath)
	}

	baseTransport := httpproxy.HTTPTransport()

	// Clone it if it's an *http.Transport, otherwise create a new one
	var transport *http.Transport
	if t, ok := baseTransport.(*http.Transport); ok {
		transport = t.Clone()
	} else {
		transport = &http.Transport{}
	}

	// Override DialContext to use Unix socket
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
	// Define the first DNS zone with specific records
	zone1 := types.Zone{
		Name: "crc.testing.",
		Records: []types.Record{
			{Name: "host", IP: net.ParseIP("192.168.127.254")},
			{Name: "api", IP: net.ParseIP("192.168.127.2")},
			{Name: "api-int", IP: net.ParseIP("192.168.127.2")},
			{Name: "crc", IP: net.ParseIP("192.168.126.11")},
		},
		DefaultIP: net.ParseIP("192.168.127.2"),
	}

	if err := gvClient.AddDNS(&zone1); err != nil {
		return errors.Wrap(err, "failed to add DNS zone to gvproxy")
	}

	// Define the second DNS zone with default IP (wildcard for apps)
	zone2 := types.Zone{
		Name:      "apps-crc.testing.",
		DefaultIP: net.ParseIP("192.168.127.2"),
	}

	// Add the second zone
	if err := gvClient.AddDNS(&zone2); err != nil {
		return errors.Wrap(err, "failed to add DNS zone to gvproxy")
	}

	logging.Info("Successfully configured internal DNS server")
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
