package adminhelper

import (
	"github.com/crc-org/admin-helper/pkg/hosts"
	"github.com/crc-org/admin-helper/pkg/types"
	"github.com/crc-org/crc/pkg/crc/constants"
)

// UpdateHostsFile updates the host's /etc/hosts file with Instance IP.
func UpdateHostsFile(instanceIP string, hostnames ...string) error {
	if err := RemoveFromHostsFile(hostnames...); err != nil {
		return err
	}
	return AddToHostsFile(instanceIP, hostnames...)
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

	return instance().Add(&types.AddRequest{
		IP:    instanceIP,
		Hosts: filtered,
	})
}

func RemoveFromHostsFile(hostnames ...string) error {
	return instance().Remove(&types.RemoveRequest{
		Hosts: hostnames,
	})
}

func CleanHostsFile() error {
	return instance().Clean(&types.CleanRequest{
		Domains: []string{constants.ClusterDomain, constants.AppsDomain},
	})
}

type helper interface {
	Add(req *types.AddRequest) error
	Remove(req *types.RemoveRequest) error
	Clean(req *types.CleanRequest) error
}
