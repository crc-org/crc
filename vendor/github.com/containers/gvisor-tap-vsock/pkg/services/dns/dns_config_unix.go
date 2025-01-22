//go:build !windows

package dns

import (
	"net"
	"net/netip"

	"github.com/containers/gvisor-tap-vsock/pkg/utils"
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
)

func (r *dnsConfig) init() error {
	if err := r.refreshNameservers(); err != nil {
		return err
	}

	w, err := utils.NewFileWatcher(etcResolvConfPath)
	if err != nil {
		return err
	}

	if err := w.Start(func() { _ = r.refreshNameservers() }); err != nil {
		return err
	}

	return nil
}

func (r *dnsConfig) refreshNameservers() error {
	nsList, err := getDNSHostAndPort(etcResolvConfPath)
	if err != nil {
		log.Errorf("can't load dns nameservers: %v", err)
		return err
	}

	log.Infof("reloading dns nameservers to %v", nsList)

	r.mu.Lock()
	r.nameservers = nsList
	r.mu.Unlock()
	return nil
}

const etcResolvConfPath = "/etc/resolv.conf"

func getDNSHostAndPort(path string) ([]string, error) {
	conf, err := dns.ClientConfigFromFile(path)
	if err != nil {
		return []string{}, err
	}
	hosts := make([]string, 0, len(conf.Servers))
	for _, server := range conf.Servers {
		dnsIP, err := netip.ParseAddr(server)
		if err != nil {
			log.Errorf("Failed to parse DNS IP address: %s", server)
			continue
		}
		// add only ipv4 dns addresses
		if dnsIP.Is4() {
			hosts = append(hosts, net.JoinHostPort(server, conf.Port))
		}
	}

	return hosts, nil
}
