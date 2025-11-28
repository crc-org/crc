package machine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	"github.com/pkg/errors"
)

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
	zoneJSON, err := json.Marshal(zoneData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal DNS zone to JSON")
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

	req, err := http.NewRequest("POST", "http://gvproxy/services/dns/add", bytes.NewBuffer(zoneJSON))
	if err != nil {
		return errors.Wrap(err, "failed to create HTTP request")
	}
	req.Header.Set("Content-Type", "application/json")

	logging.Debugf("Adding DNS zone to gvproxy: %s", string(zoneJSON))

	resp, err := client.Do(req) //nolint:gosec
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
	uid := os.Getuid()
	socketPath := filepath.Join("/run", "user", strconv.Itoa(uid), "macadam", "qemu", fmt.Sprintf("%s-gvproxy-api.sock", vmName))

	if _, err := os.Stat(socketPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("gvproxy socket does not exist at %s", socketPath)
		}
		return errors.Wrapf(err, "failed to check gvproxy socket at %s", socketPath)
	}

	zone1 := DNSZone{
		Name: "crc.testing.",
		Records: []DNSRecord{
			{Name: "host", IP: "192.168.127.254"},
			{Name: "api", IP: "192.168.127.2"},
			{Name: "api-int", IP: "192.168.127.2"},
			{Name: "crc", IP: "192.168.126.11"},
		},
	}

	if err := addDNSZoneToGVProxy(socketPath, zone1); err != nil {
		return errors.Wrap(err, "failed to add crc.testing zone")
	}

	zone2 := DNSZoneWithDefault{
		Name:      "apps-crc.testing.",
		DefaultIP: "192.168.127.2",
	}

	if err := addDNSZoneToGVProxy(socketPath, zone2); err != nil {
		return errors.Wrap(err, "failed to add apps-crc.testing zone")
	}

	logging.Info("Successfully configured internal DNS server")
	return nil
}
