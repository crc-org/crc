//go:build windows

package dns

import (
	"net/netip"
	"strconv"

	qdmDns "github.com/qdm12/dns/v2/pkg/nameserver"
)

func getDNSHostAndPort() (string, string, error) {
	nameservers := qdmDns.GetDNSServers()

	var nameserver netip.AddrPort
	for _, n := range nameservers {
		// return first non ipv6 nameserver
		if n.Addr().Is4() {
			nameserver = n
			break
		}
	}

	return nameserver.Addr().String(), strconv.Itoa(int(nameserver.Port())), nil

}
