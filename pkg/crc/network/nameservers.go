package network

import (
	"fmt"
	"os"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/crc/systemd"
	"github.com/crc-org/crc/v2/pkg/crc/systemd/states"
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

func UpdateResolvFileOnInstance(sshRunner *ssh.Runner, resolvFileValues ResolvFileValues) error {
	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	// Check if ovs-configuration.service exist and if not then it is old bundle and use the same way to
	// update resolve.conf file
	if state, err := sd.Status("ovs-configuration.service"); err != nil || state == states.NotFound {
		if err := replaceResolvConfFile(sshRunner, resolvFileValues); err != nil {
			return fmt.Errorf("error updating resolv.conf file: %s", err)
		}
		return nil
	}

	if err := sd.Start("ovs-configuration.service"); err != nil {
		return err
	}

	return updateNetworkManagerConfig(sd, sshRunner, resolvFileValues)
}

func replaceResolvConfFile(sshRunner *ssh.Runner, resolvFileValues ResolvFileValues) error {
	resolvFile, err := CreateResolvFile(resolvFileValues)
	if err != nil {
		return fmt.Errorf("error to create resolv conf file: %v", err)
	}
	err = sshRunner.CopyDataPrivileged([]byte(resolvFile), "/etc/resolv.conf", 0644)
	if err != nil {
		return fmt.Errorf("error creating /etc/resolv on instance: %s", err.Error())
	}
	return nil
}

func updateNetworkManagerConfig(sd *systemd.Commander, sshRunner *ssh.Runner, resolvFileValues ResolvFileValues) error {
	nameservers := strings.Join(resolvFileValues.GetNameServer(), ",")
	searchDomains := strings.Join(resolvFileValues.GetSearchDomains(), ",")
	// When ovs-configuration service is running, name of the connection should be ovs-if-br-ex
	_, stderr, err := sshRunner.RunPrivileged("Update resolv.conf file", "nmcli", "con", "modify", "--temporary", "ovs-if-br-ex",
		"ipv4.dns", nameservers, "ipv4.dns-search", searchDomains)
	if err != nil {
		return fmt.Errorf("failed to update resolv.conf file %s: %v", stderr, err)
	}
	return sd.Restart("NetworkManager.service")
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
	out, err := os.ReadFile("/etc/resolv.conf")
	/*
		if crcos.CurrentOS() == crcos.WINDOWS {
			// TODO: we need to add logic in case of windows.
		}
	*/

	if err != nil {
		return nil, fmt.Errorf("failed to read resolv.conf: %v", err)
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
