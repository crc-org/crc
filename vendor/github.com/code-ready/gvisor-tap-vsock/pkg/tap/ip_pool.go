package tap

import (
	"errors"
	"net"
	"sync"

	"github.com/apparentlymart/go-cidr/cidr"
)

type IPPool struct {
	base   *net.IPNet
	count  uint64
	leases map[string]int
	lock   sync.Mutex
}

func NewIPPool(base *net.IPNet) *IPPool {
	return &IPPool{
		base:   base,
		count:  cidr.AddressCount(base),
		leases: make(map[string]int),
	}
}

func (p *IPPool) Leases() map[string]int {
	return p.leases
}

func (p *IPPool) Mask() int {
	ones, _ := p.base.Mask.Size()
	return ones
}

func (p *IPPool) Assign(id int) (net.IP, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	var i uint64
	for i = 1; i < p.count; i++ {
		candidate, err := cidr.Host(p.base, int(i))
		if err != nil {
			continue
		}
		if _, ok := p.leases[candidate.String()]; !ok {
			p.leases[candidate.String()] = id
			return candidate, nil
		}
	}
	return nil, errors.New("cannot find available IP")
}

func (p *IPPool) Reserve(ip net.IP, id int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.leases[ip.String()] = id
}

func (p *IPPool) Release(given int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	var found string
	for ip, id := range p.leases {
		if id == given {
			found = ip
			break
		}
	}
	if found != "" {
		delete(p.leases, found)
	}
}
