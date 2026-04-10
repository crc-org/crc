//go:build !windows

package test_test

import (
	crcos "github.com/crc-org/crc/v2/pkg/os"
)

// addHostsEntry adds an entry to /etc/hosts using sudo tee.
func addHostsEntry(ip, hostname string) error {
	cmd := `printf '%s %s\n' "$1" "$2" | tee -a /etc/hosts > /dev/null`
	_, _, err := crcos.RunPrivileged("adding hosts entry for proxy test", "sh", "-c", cmd, "addHostsEntry", ip, hostname)
	return err
}
