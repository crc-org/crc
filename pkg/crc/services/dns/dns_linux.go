package dns

import (
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/services"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig) error {
	// We might need to set the firewall here to forward
	// Update /etc/hosts file for host
	if serviceConfig.ModifyHostsFile {
		return addOpenShiftHosts(serviceConfig)
	} else {
		logging.Infof("Skipping hosts file modification")
	}

	return nil
}
