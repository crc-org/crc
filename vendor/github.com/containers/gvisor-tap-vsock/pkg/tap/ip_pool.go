package tap

import (
	"errors"
	"net/netip"
	"sync"
)

type IPPool struct {
	base   netip.Prefix
	leases map[netip.Addr]string
	lock   sync.Mutex
}

func NewIPPool(subnet string) (*IPPool, error) {
	base, err := netip.ParsePrefix(subnet)
	if err != nil {
		return nil, err
	}
	return &IPPool{
		base:   base,
		leases: make(map[netip.Addr]string),
	}, nil
}

func (p *IPPool) Leases() map[string]string {
	p.lock.Lock()
	defer p.lock.Unlock()
	leases := map[string]string{}
	for ip, mac := range p.leases {
		leases[ip.String()] = mac
	}
	return leases
}

func (p *IPPool) Mask() int {
	return p.base.Bits()
}

func (p *IPPool) GetOrAssign(mac string) (netip.Addr, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for ip, candidate := range p.leases {
		if candidate == mac {
			return ip, nil
		}
	}

	// Start from the first usable IP (network address + 1)
	candidate := p.base.Masked().Addr().Next()

	// Iterate through all IPs in the subnet
	for candidate.IsValid() && p.base.Contains(candidate) {
		if _, ok := p.leases[candidate]; !ok {
			p.leases[candidate] = mac
			return candidate, nil
		}
		candidate = candidate.Next()
	}

	return netip.Addr{}, errors.New("cannot find available IP")
}

func (p *IPPool) Reserve(ip string, mac string) error {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return err
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	p.leases[addr] = mac
	return nil
}

func (p *IPPool) Release(given string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for ip, mac := range p.leases {
		if mac == given {
			delete(p.leases, ip)
			break
		}
	}
}
