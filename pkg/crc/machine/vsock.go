package machine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	crcErrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
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

// DNSRecord represents a DNS record entry
type DNSRecord struct {
	Name string `json:"Name"`
	IP   string `json:"IP"`
}

// DNSZone represents a DNS zone configuration with specific records
type DNSZone struct {
	Name    string      `json:"Name"`
	Records []DNSRecord `json:"Records,omitempty"`
}

// DNSZoneWithDefault represents a DNS zone configuration with a default IP (wildcard)
type DNSZoneWithDefault struct {
	Name      string `json:"Name"`
	DefaultIP string `json:"DefaultIP"`
}

// addDNSZoneToGVProxy adds a DNS zone to gvproxy via Unix socket
func addDNSZoneToGVProxy(socketPath string, zoneData interface{}) error {
	// Marshal the zone to JSON
	zoneJSON, err := json.Marshal(zoneData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal DNS zone to JSON")
	}

	// Create HTTP client with Unix socket transport and proxy support
	// Get the base transport from httpproxy package which handles proxy configuration
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

	// Make POST request to gvproxy DNS service
	req, err := http.NewRequest("POST", "http://gvproxy/services/dns/add", bytes.NewBuffer(zoneJSON))
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}
	req.Header.Set("Content-Type", "application/json")

	logging.Debugf("Adding DNS zone to gvproxy: %s", string(zoneJSON))

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to add DNS zone via gvproxy socket %s", socketPath)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("gvproxy DNS service returned status %d", resp.StatusCode)
	}

	return nil
}

// enableInternalDNS configures the internal DNS server via gvproxy Unix socket
func enableInternalDNS(vmName string, bundleInfo interface{}) error {
	// Construct the gvproxy socket path
	// Path format: /run/user/<uid>/macadam/qemu/<vmname>-gvproxy-api.sock
	uid := os.Getuid()
	socketPath := filepath.Join("/run", "user", strconv.Itoa(uid), "macadam", "qemu", fmt.Sprintf("%s-gvproxy-api.sock", vmName))

	// Check if the socket path exists
	if _, err := os.Stat(socketPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("gvproxy socket does not exist at %s", socketPath)
		}
		return errors.Wrapf(err, "failed to check gvproxy socket at %s", socketPath)
	}

	// Define the first DNS zone with specific records
	zone1 := DNSZone{
		Name: "crc.testing.",
		Records: []DNSRecord{
			{Name: "host", IP: "192.168.127.254"},
			{Name: "api", IP: "192.168.127.2"},
			{Name: "api-int", IP: "192.168.127.2"},
			{Name: "crc", IP: "192.168.126.11"},
		},
	}

	// Add the first zone
	if err := addDNSZoneToGVProxy(socketPath, zone1); err != nil {
		return errors.Wrap(err, "failed to add crc.testing zone")
	}

	// Define the second DNS zone with default IP (wildcard for apps)
	zone2 := DNSZoneWithDefault{
		Name:      "apps-crc.testing.",
		DefaultIP: "192.168.127.2",
	}

	// Add the second zone
	if err := addDNSZoneToGVProxy(socketPath, zone2); err != nil {
		return errors.Wrap(err, "failed to add apps-crc.testing zone")
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
