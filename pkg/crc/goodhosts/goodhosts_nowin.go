// +build !windows

package goodhosts

import (
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcos "github.com/code-ready/crc/pkg/os"
)

var (
	goodhostPath = filepath.Join(constants.CrcBinDir, constants.GoodhostsBinaryName)
)

// UpdateHostsFile updates the host's /etc/hosts file with Instance IP.
func UpdateHostsFile(instanceIP string, hostnames ...string) error {
	if err := RemoveFromHostsFile(instanceIP, hostnames...); err != nil {
		return err
	}
	if err := AddToHostsFile(instanceIP, hostnames...); err != nil {
		return err
	}
	return nil
}

func AddToHostsFile(instanceIP string, hostnames ...string) error {
	for _, hostname := range hostnames {
		if _, _, err := crcos.RunWithDefaultLocale(goodhostPath, "add", instanceIP, hostname); err != nil {
			return err
		}
	}
	return nil
}

func RemoveFromHostsFile(instanceIP string, hostnames ...string) error {
	// If only instanceIP provided then remove all the entry from that instance IP
	if len(hostnames) == 0 {
		if _, _, err := crcos.RunWithDefaultLocale(goodhostPath, "rm", instanceIP); err != nil {
			return err
		}
		return nil
	}
	for _, hostname := range hostnames {
		if _, _, err := crcos.RunWithDefaultLocale(goodhostPath, "rm", hostname); err != nil {
			return err
		}
	}
	return nil
}
