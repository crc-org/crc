package network

import (
	"fmt"
	"strings"

	"github.com/code-ready/crc/pkg/os/windows/powershell"
)

const hypervDefaultVirtualSwitchId = "c08cb7b8-9b3c-408e-8e30-5e16a3aeb444"

func GetDefaultSwitchName() (bool, string) {
	getDefaultSwitchNameCmd := fmt.Sprintf("[Console]::OutputEncoding = [Text.Encoding]::UTF8; Get-VMSwitch -Id %s | ForEach-Object { $_.Name }", hypervDefaultVirtualSwitchId)
	stdOut, stdErr, _ := powershell.Execute(getDefaultSwitchNameCmd)

	// If stdErr contains the command then execution failed
	if strings.Contains(stdErr, "Get-VMSwitch") {
		return false, ""
	}

	return true, strings.TrimSpace(stdOut)
}
