//go:build windows

package dns

import (
	"net"
	"strconv"

	qdmDns "github.com/qdm12/dns/v2/pkg/nameserver"
)

func getDNSHostAndPort() ([]string, error) {
	nameservers := qdmDns.GetDNSServers()

	var dnsServers []string
	for _, n := range nameservers {
		// return only ipv4 nameservers
		if n.Addr().Is4() {
			dnsServers = append(dnsServers, net.JoinHostPort(n.Addr().String(), strconv.Itoa(int(n.Port()))))
		}
	}

	return dnsServers, nil

}
