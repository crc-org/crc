package dns

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/services"

	winnet "github.com/code-ready/crc/pkg/os/windows/network"
	"github.com/code-ready/crc/pkg/os/windows/powershell"
	"github.com/code-ready/crc/pkg/os/windows/win32"
)

const (
	// Alternative
	AlternativeNetwork = "crc"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// bailout for Virtualbox
	if serviceConfig.DriverName == "virtualbox" {
		output.Outln("Please follow instructions in the documentation about setting hostnames for Virtualbox.")
		result.Success = true
		return *result, nil
	}

	_, switchName := winnet.SelectSwitchByNameOrDefault(AlternativeNetwork)
	networkInterface := fmt.Sprintf("vEthernet (%s)", switchName)

	setInterfaceNameserverValue(networkInterface, serviceConfig.IP)

	time.Sleep(2 * time.Second)

	if !contains(getInterfaceNameserverValues(networkInterface), serviceConfig.IP) {
		err := errors.New("Nameserver not successfully set")
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	result.Success = true
	return *result, nil
}

func getInterfaceNameserverValues(iface string) []string {
	getDNSServerCommand := fmt.Sprintf(`(Get-DnsClientServerAddress "%s")[0].ServerAddresses`, iface)
	stdOut, _, _ := powershell.Execute(getDNSServerCommand)

	return parseLines(stdOut)
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func setInterfaceNameserverValue(iface string, address string) {
	exe := "netsh"
	args := fmt.Sprintf(`interface ip set dns "%s" static %s primary`, iface, address)

	// ignore the error as this is useless (prefer not to use nolint here)
	_ = win32.ShellExecuteAsAdmin(fmt.Sprintf("add dns server address to interface %s", iface), win32.HWND_DESKTOP, exe, args, "", 0)
}

func parseLines(input string) []string {
	output := []string{}

	s := bufio.NewScanner(strings.NewReader(input))
	for s.Scan() {
		output = append(output, s.Text())
	}

	return output
}
