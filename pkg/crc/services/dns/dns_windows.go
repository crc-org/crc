package dns

import (
	"fmt"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/services"
	winnet "github.com/crc-org/crc/v2/pkg/os/windows/network"
	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
	"github.com/crc-org/crc/v2/pkg/os/windows/win32"
	crcstrings "github.com/crc-org/crc/v2/pkg/strings"
)

const (
	// Alternative
	AlternativeNetwork = "crc"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig) error {
	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		return addOpenShiftHosts(serviceConfig)
	}

	_, switchName := winnet.SelectSwitchByNameOrDefault(AlternativeNetwork)
	networkInterface := fmt.Sprintf("vEthernet (%s)", switchName)

	setInterfaceNameserverValue(networkInterface, serviceConfig.IP)

	time.Sleep(2 * time.Second)

	if !crcstrings.Contains(getInterfaceNameserverValues(networkInterface), serviceConfig.IP) {
		return fmt.Errorf("Nameserver %s not successfully set on interface %s. Perhaps you can try this new network mode: https://github.com/crc-org/crc/wiki/VPN-support--with-an--userland-network-stack", serviceConfig.IP, networkInterface)
	}
	return nil
}

func getInterfaceNameserverValues(iface string) []string {
	getDNSServerCommand := fmt.Sprintf(`(Get-DnsClientServerAddress "%s")[0].ServerAddresses`, iface)
	stdOut, _, _ := powershell.Execute(getDNSServerCommand)

	return crcstrings.SplitLines(stdOut)
}

func setInterfaceNameserverValue(iface string, address string) {
	exe := "netsh"
	args := fmt.Sprintf(`interface ip set dns "%s" static %s primary`, iface, address)

	// ignore the error as this is useless (prefer not to use nolint here)
	_ = win32.ShellExecuteAsAdmin(fmt.Sprintf("add dns server address to interface %s", iface), win32.HwndDesktop, exe, args, "", 0)
}
