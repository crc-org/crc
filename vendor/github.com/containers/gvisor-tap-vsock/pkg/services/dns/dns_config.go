package dns

import "sync"

type dnsConfig struct {
	mu          sync.RWMutex
	nameservers []string
}

func newDNSConfig() (*dnsConfig, error) {
	r := &dnsConfig{nameservers: []string{}}
	if err := r.init(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *dnsConfig) Nameservers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.nameservers
}
