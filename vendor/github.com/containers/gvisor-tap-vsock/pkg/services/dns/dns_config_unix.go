//go:build !windows

package dns

import (
	"github.com/miekg/dns"
)

func getDNSHostAndPort() (string, string, error) {
	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return "", "", err
	}
	// TODO: use all configured nameservers, instead just first one
	nameserver := conf.Servers[0]

	return nameserver, conf.Port, nil
}
