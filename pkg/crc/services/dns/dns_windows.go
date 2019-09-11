package dns

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/services"

	"github.com/code-ready/crc/pkg/os/windows/powershell"
	"github.com/code-ready/crc/pkg/os/windows/win32"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// TODO: localize
	networkInterface := "vEthernet (Default Switch)" //getMainInterface()

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

func formatValues(serverAddresses []string) string {
	var out string
	for index, serverAddress := range serverAddresses {
		out = fmt.Sprintf(`%s"%s"`, out, serverAddress)
		if index < len(serverAddresses)-1 {
			out = fmt.Sprintf(`%s, `, out)
		}
	}

	return out
}

func setInterfaceNameserverValue(iface string, address string) {
	exe := "netsh"
	args := fmt.Sprintf(`interface ip set dns "%s" static %s primary`, iface, address)

	win32.ShellExecuteAsAdmin(fmt.Sprintf("add dns server address to interface %s", iface), win32.HWND_DESKTOP, exe, args, "", 0)
}

func getMainInterface() string {
	getMainInterfaceCommand := `(Get-NetAdapter | Where-Object {$_.MediaConnectionState -eq 'Connected'} | Sort-Object LinkSpeed -Descending)[0].Name`
	mainInterface, _, _ := powershell.Execute(getMainInterfaceCommand)

	return strings.TrimSpace(mainInterface)
}

func parseLines(input string) []string {
	output := []string{}

	s := bufio.NewScanner(strings.NewReader(input))
	for s.Scan() {
		output = append(output, s.Text())
	}

	return output
}
