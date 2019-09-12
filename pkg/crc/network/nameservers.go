package network

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/code-ready/crc/pkg/crc/ssh"
)

// HasGivenNameserversConfigured returns true if the instance uses a provided nameserver.
func HasGivenNameserversConfigured(sshRunner *ssh.SSHRunner, nameserver NameServer) (bool, error) {
	cmd := "cat /etc/resolv.conf"
	out, err := sshRunner.Run(cmd)

	if err != nil {
		return false, err
	}

	return strings.Contains(out, nameserver.IPAddress), nil
}

func GetResolvValuesFromInstance(sshRunner *ssh.SSHRunner) (*ResolvFileValues, error) {
	cmd := "cat /etc/resolv.conf"
	out, err := sshRunner.Run(cmd)

	if err != nil {
		return nil, err
	}

	return parseResolveConfFile(out)
}

func CreateResolvFileOnInstance(sshRunner *ssh.SSHRunner, resolvFileValues ResolvFileValues) {
	resolvFile, _ := CreateResolvFile(resolvFileValues)
	encodedFile := base64.StdEncoding.EncodeToString([]byte(resolvFile))

	executeCommandOrExit(sshRunner,
		fmt.Sprintf("echo %s | base64 --decode | sudo tee /etc/resolv.conf > /dev/null", encodedFile),
		"Error creating /etc/resolv on instance")
}

// AddNameserversToInstance will add additional nameservers to the end of the
// /etc/resolv.conf file inside the instance.
func AddNameserversToInstance(sshRunner *ssh.SSHRunner, nameservers []NameServer) {
	for _, ns := range nameservers {
		addNameserverToInstance(sshRunner, ns)
	}
}

// writes nameserver to the /etc/resolv.conf inside the instance
func addNameserverToInstance(sshRunner *ssh.SSHRunner, nameserver NameServer) {
	executeCommandOrExit(sshRunner,
		fmt.Sprintf("NS=%s; cat /etc/resolv.conf |grep -i \"^nameserver $NS\" || echo \"nameserver $NS\" | sudo tee -a /etc/resolv.conf", nameserver.IPAddress),
		"Error adding nameserver")
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
