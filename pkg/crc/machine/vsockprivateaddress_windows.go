package machine

import (
	"fmt"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
)

const (
	fallbackPrivateAddress = "127.0.0.1"
	adapterName            = "vEthernet (WSL)"
)

// On Windows it would be useful if WSL2 can access VSock resources as well
// This address should work on both the windows side and the WSL2 side
func VsockPrivateAddress() (addrStr string) {
	if !constants.WSLNetworkAccess {
		return fallbackPrivateAddress
	}
	cmd := fmt.Sprintf(`(Get-NetIPAddress -AddressFamily IPv4 -InterfaceAlias "%s").IPAddress`, adapterName)
	stdout, stderr, err := powershell.Execute(cmd)
	if err != nil {
		logging.Debugf("Unable to find IP address for WSL vswitch: %v: %s", err, stderr)
		return fallbackPrivateAddress
	}
	if strings.TrimSpace(stdout) == "" {
		return fallbackPrivateAddress
	}
	return strings.TrimSpace(stdout)
}
