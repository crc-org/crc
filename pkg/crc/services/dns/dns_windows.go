package dns

import (
	"fmt"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/services"

	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// NOTE: this is very Hyper-V specific
	// TODO: localize the use of the Default Switch
	setDNSServerCommand := fmt.Sprintf(`Set-DnsClientServerAddress "vEthernet (Default Switch)" -ServerAddress ("%s")`, serviceConfig.IP)
	powershell.ExecuteAsAdmin(setDNSServerCommand)

	time.Sleep(2 * time.Second)

	getDNSServerCommand := `(Get-DnsClientServerAddress "vEthernet (Default Switch)").ServerAddresses`
	stdOut, _, _ := powershell.Execute(getDNSServerCommand)

	if !strings.Contains(stdOut, serviceConfig.IP) {
		err := errors.New("Nameserver not successfully set")
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	result.Success = true
	return *result, nil
}
