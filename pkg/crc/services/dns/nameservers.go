package dns

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/code-ready/crc/pkg/crc/ssh"
)

// HasGivenNameserversConfigured returns true if the instance uses a provided nameserver.
func HasGivenNameserversConfigured(sshRunner *ssh.Runner, nameserver NameServer) (bool, error) {
	cmd := "cat /etc/resolv.conf"
	out, _, err := sshRunner.Run(cmd)

	if err != nil {
		return false, err
	}

	return strings.Contains(out, nameserver.IPAddress), nil
}

func GetResolvValuesFromInstance(sshRunner *ssh.Runner) (*ResolvFileValues, error) {
	cmd := "cat /etc/resolv.conf"
	out, _, err := sshRunner.Run(cmd)

	if err != nil {
		return nil, err
	}

	return parseResolveConfFile(out)
}

func CreateResolvFileOnInstance(serviceConfig ServicePostStartConfig) error {
	resolvFileValues, err := getResolvFileValues(serviceConfig)
	if err != nil {
		return err
	}
	// override resolv.conf file

	resolvFile, _ := CreateResolvFile(resolvFileValues)

	err = serviceConfig.SSHRunner.CopyData([]byte(resolvFile), "/etc/resolv.conf", 0644)
	if err != nil {
		return fmt.Errorf("Error creating /etc/resolv on instance: %s", err.Error())
	}
	return nil
}

// AddNameserversToInstance will add additional nameservers to the end of the
// /etc/resolv.conf file inside the instance.
func AddNameserversToInstance(sshRunner *ssh.Runner, nameservers []NameServer) error {
	for _, ns := range nameservers {
		if err := addNameserverToInstance(sshRunner, ns); err != nil {
			return err
		}
	}
	return nil
}

// writes nameserver to the /etc/resolv.conf inside the instance
func addNameserverToInstance(sshRunner *ssh.Runner, nameserver NameServer) error {
	_, _, err := sshRunner.Run(fmt.Sprintf("NS=%s; cat /etc/resolv.conf |grep -i \"^nameserver $NS\" || echo \"nameserver $NS\" | sudo tee -a /etc/resolv.conf", nameserver.IPAddress))
	if err != nil {
		return fmt.Errorf("%s: %s", "Error adding nameserver", err.Error())
	}
	return nil
}

func GetResolvValuesFromHost() (*ResolvFileValues, error) {
	// TODO: we need to add runtime OS in case of windows.
	out, err := ioutil.ReadFile("/etc/resolv.conf")
	/*
		if crcos.CurrentOS() == crcos.WINDOWS {
			// TODO: we need to add logic in case of windows.
		}
	*/

	if err != nil {
		return nil, fmt.Errorf("Failed to read resolv.conf: %v", err)
	}
	return parseResolveConfFile(string(out))
}

func parseResolveConfFile(resolvfile string) (*ResolvFileValues, error) {
	searchdomains := []SearchDomain{}
	nameservers := []NameServer{}

	for _, line := range strings.Split(strings.TrimSuffix(resolvfile, "\n"), "\n") {
		if len(line) > 0 && (line[0] == ';' || line[0] == '#') {
			continue
		}

		f := strings.Fields(line)
		if len(f) < 1 {
			continue
		}

		switch f[0] {
		case "nameserver":
			nameservers = append(nameservers, NameServer{IPAddress: f[1]})
		case "search":
			for i := 0; i < len(f)-1; i++ {
				searchdomains = append(searchdomains, SearchDomain{Domain: f[i+1]})
			}
		default:
			// ignore
		}
	}

	resolvFileValues := ResolvFileValues{
		SearchDomains: searchdomains,
		NameServers:   nameservers,
	}

	return &resolvFileValues, nil
}
