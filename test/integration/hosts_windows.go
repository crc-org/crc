//go:build windows

package test_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/os/windows/powershell"
)

// addHostsEntry adds an entry to the Windows hosts file using ExecuteAsAdmin.
func addHostsEntry(ip, hostname string) error {
	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = `C:\Windows`
	}
	hostsPath := filepath.Join(systemRoot, "System32", "drivers", "etc", "hosts")

	entry := fmt.Sprintf("%s %s", ip, hostname)
	cmd := fmt.Sprintf(`Add-Content -Path "%s" -Value "%s"`, hostsPath, entry)

	_, _, err := powershell.ExecuteAsAdmin("adding hosts entry for proxy test", cmd)
	return err
}
