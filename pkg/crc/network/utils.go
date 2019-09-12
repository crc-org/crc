package network

import (
	"fmt"
	"net"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/ssh"
)

func executeCommandOrExit(sshRunner *ssh.SSHRunner, command string, errorMessage string) string {
	result, err := sshRunner.Run(command)

	if err != nil {
		errors.ExitWithMessage(1, fmt.Sprintf("%s: %s", errorMessage, err.Error()))
	}
	return result
}

// NetworkContains returns true if the IP address belongs to the network given
func NetworkContains(network string, ip string) bool {
	_, ipnet, _ := net.ParseCIDR(network)
	address := net.ParseIP(ip)
	return ipnet.Contains(address)
}

// HostIPs returns the IP addresses assigned to the host
func HostIPs() []string {
	ips := []string{}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			ips = append(ips, addr.String())
		}
	}

	return ips
}

func DetermineHostIP(instanceIP string) (string, error) {
	for _, hostaddr := range HostIPs() {

		if NetworkContains(hostaddr, instanceIP) {
			hostip, _, _ := net.ParseCIDR(hostaddr)
			// This step is not working with Windows + VirtualBox as of now
			// This test is required for CIFS mount-folder case.
			// Details: https://github.com/minishift/minishift/issues/2561
			/*if IsIPReachable(driver, hostip.String(), false) {
				return hostip.String(), nil
			}*/
			return hostip.String(), nil
		}
	}

	return "", errors.New("unknown error occurred")
}
