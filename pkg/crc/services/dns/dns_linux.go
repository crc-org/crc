package dns

import (
	"github.com/crc-org/crc/v2/pkg/crc/services"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig) error {
	// We might need to set the firewall here to forward
	// Update /etc/hosts file for host
	return addOpenShiftHosts(serviceConfig)
}
