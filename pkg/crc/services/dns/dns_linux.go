package dns

import (
	"github.com/code-ready/crc/pkg/crc/services"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig) error {
	// We might need to set the firewall here to forward
	// Update /etc/hosts file for host
	return addOpenShiftHosts(serviceConfig)
}
