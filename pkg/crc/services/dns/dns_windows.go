package dns

import (
	"github.com/code-ready/crc/pkg/crc/services"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	result.Success = true
	return *result, nil
}
