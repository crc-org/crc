package adminhelper

import (
	"path/filepath"

	"github.com/code-ready/admin-helper/pkg/hosts"
	"github.com/code-ready/crc/pkg/crc/constants"
)

var (
	adminHelperPath = filepath.Join(constants.CrcBinDir, constants.AdminHelperExecutableName)
)

// UpdateHostsFile updates the host's /etc/hosts file with Instance IP.
func UpdateHostsFile(instanceIP string, hostnames ...string) error {
	if err := RemoveFromHostsFile(hostnames...); err != nil {
		return err
	}
	if err := AddToHostsFile(instanceIP, hostnames...); err != nil {
		return err
	}
	return nil
}

func AddToHostsFile(instanceIP string, hostnames ...string) error {
	hosts, err := hosts.New()
	if err != nil {
		return err
	}
	var filtered []string
	for _, hostname := range hostnames {
		if !hosts.Contains(instanceIP, hostname) {
			filtered = append(filtered, hostname)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return execute(append([]string{"add", instanceIP}, hostnames...)...)
}

func RemoveFromHostsFile(hostnames ...string) error {
	return execute(append([]string{"rm"}, hostnames...)...)
}

func CleanHostsFile() error {
	return execute([]string{"clean", constants.ClusterDomain, constants.AppsDomain}...)
}
