package dns

import (
	"github.com/code-ready/crc/pkg/crc/services"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// We might need to set the firewall here to forward

	result.Success = true
	return *result, nil
}
