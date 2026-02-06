// Package client provides go bindings for gvisor-tap-vsock HTTP API
// This API is accessible over the endpoint specified with the `--services` or `--listen` arguments to `gvproxy`
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/containers/gvisor-tap-vsock/pkg/types"
)

type Client struct {
	client *http.Client
	base   string
}

// New returns a new instance of a Client. client will be used for the HTTP communication, and base specifies the base path the HTTP API is available at.
func New(client *http.Client, base string) *Client {
	return &Client{
		client: client,
		base:   base,
	}
}

// List lists all the forwarded ports between host and guest
//
// Request:
// GET /services/forwarder/all
// Response:
// [{"local":"127.0.0.1:2223","remote":"192.168.127.2:22","protocol":"tcp"}]
func (c *Client) List() ([]types.ExposeRequest, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, "/services/forwarder/all"))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}
	dec := json.NewDecoder(res.Body)
	var ports []types.ExposeRequest
	if err := dec.Decode(&ports); err != nil {
		return nil, err
	}
	return ports, nil
}

// Expose forwards a new port between host and guest
//
// Request:
// POST /services/forwarder/expose
// {"local":"127.0.0.1:2224","remote":"192.168.127.2:22","protocol":"tcp"}
// Response:
// HTTP Status Code
func (c *Client) Expose(req *types.ExposeRequest) error {
	bin, err := json.Marshal(req)
	if err != nil {
		return err
	}
	res, err := c.client.Post(fmt.Sprintf("%s%s", c.base, "/services/forwarder/expose"), "application/json", bytes.NewReader(bin))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("error while reading error message: %v", readErr)
		}
		return errors.New(strings.TrimSpace(string(err)))
	}
	return nil
}

// Unexpose stops forwarding a port between host and guest
//
// Request:
// POST /services/forwarder/unexpose
// {"local":"127.0.0.1:2224","protocol":"tcp"}
// Response:
// HTTP Status Code
func (c *Client) Unexpose(req *types.UnexposeRequest) error {
	bin, err := json.Marshal(req)
	if err != nil {
		return err
	}
	res, err := c.client.Post(fmt.Sprintf("%s%s", c.base, "/services/forwarder/unexpose"), "application/json", bytes.NewReader(bin))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("error while reading error message: %v", readErr)
		}
		return errors.New(strings.TrimSpace(string(err)))
	}
	return nil
}

// ListDNS shows the configuration of the built-in DNS server
//
// Request:
// GET /services/dns/all
// Response:
// [{"Name":"containers.internal.","Records":[{"Name":"gateway","IP":"192.168.127.1","Regexp":null},{"Name":"host","IP":"192.168.127.254","Regexp":null}],"DefaultIP":""},{"Name":"docker.internal.","Records":[{"Name":"gateway","IP":"192.168.127.1","Regexp":null},{"Name":"host","IP":"192.168.127.254","Regexp":null}],"DefaultIP":""}]
func (c *Client) ListDNS() ([]types.Zone, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, "/services/dns/all"))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}
	dec := json.NewDecoder(res.Body)
	var dnsZone []types.Zone
	if err := dec.Decode(&dnsZone); err != nil {
		return nil, err
	}
	return dnsZone, nil
}

// AddDNS adds a new DNS zone to the built-in DNS server
//
// Request:
// POST /services/dns/add
// {"Name":"test.internal.","Records":[{"Name":"gateway","IP":"192.168.127.1"}]}
// Response:
// HTTP Status Code
func (c *Client) AddDNS(req *types.Zone) error {
	bin, err := json.Marshal(req)
	if err != nil {
		return err
	}
	res, err := c.client.Post(fmt.Sprintf("%s%s", c.base, "/services/dns/add"), "application/json", bytes.NewReader(bin))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return fmt.Errorf("error while reading error message: %v", readErr)
		}
		return errors.New(strings.TrimSpace(string(err)))
	}
	return nil
}

// ListDHCPLeases shows the configuration of the built-in DNS server
//
// Request:
// GET /services/dhcp/leases
// Response:
// {"192.168.127.1":"5a:94:ef:e4:0c:dd","192.168.127.2":"5a:94:ef:e4:0c:ee"}
func (c *Client) ListDHCPLeases() (map[string]string, error) {
	res, err := c.client.Get(fmt.Sprintf("%s%s", c.base, "/services/dhcp/leases"))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}
	dec := json.NewDecoder(res.Body)
	var leases map[string]string
	if err := dec.Decode(&leases); err != nil {
		return nil, err
	}
	return leases, nil
}
