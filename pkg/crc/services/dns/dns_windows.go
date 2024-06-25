package dns

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/services"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig) error {
	if serviceConfig.NetworkMode != network.UserNetworkingMode {
		return fmt.Errorf("only user-mode networking is supported on Windows")
	}
	return addOpenShiftHosts(serviceConfig)
}
