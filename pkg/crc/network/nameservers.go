package network

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/machine/libmachine/drivers"
)

// HasNameserversConfigured returns true if the instance uses nameservers
// This is related to an issues when LCOW is used on Windows.
func HasNameserversConfigured(driver drivers.Driver) bool {
	cmd := "cat /etc/resolv.conf | grep -i '^nameserver' | wc -l | tr -d '\n'"
	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return false
	}

	i, _ := strconv.Atoi(out)

	return i != 0
}

func GetResolvValuesFromInstance(driver drivers.Driver) (*ResolvFileValues, error) {
	cmd := "cat /etc/resolv.conf"
	out, err := drivers.RunSSHCommandFromDriver(driver, cmd)

	if err != nil {
		return nil, err
	}

	return parseResolveConfFile(out)
}

func CreateResolvFileOnInstance(driver drivers.Driver, resolvFileValues ResolvFileValues) {
	resolvFile, _ := CreateResolvFile(resolvFileValues)
	encodedFile := base64.StdEncoding.EncodeToString([]byte(resolvFile))

	executeCommandOrExit(driver,
		fmt.Sprintf("echo %s | base64 --decode | sudo tee /etc/resolv.conf > /dev/null", encodedFile),
		"Error creating /etc/resolv on instance")
}

// AddNameserversToInstance will add additional nameservers to the end of the
// /etc/resolv.conf file inside the instance.
func AddNameserversToInstance(driver drivers.Driver, nameservers []NameServer) {
	// TODO: verify values to be valid

	for _, ns := range nameservers {
		addNameserverToInstance(driver, ns)
	}
}

// writes nameserver to the /etc/resolv.conf inside the instance
func addNameserverToInstance(driver drivers.Driver, nameserver NameServer) {
	executeCommandOrExit(driver,
		fmt.Sprintf("NS=%s; cat /etc/resolv.conf |grep -i \"^nameserver $NS\" || echo \"nameserver $NS\" | sudo tee -a /etc/resolv.conf", nameserver.IPAddress),
		"Error adding nameserver")
}

func HasNameserverConfiguredLocally(nameserver NameServer) (bool, error) {
	file, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return false, err
	}

	return strings.Contains(string(file), nameserver.IPAddress), nil
}

func GetResolvValuesFromHost() (*ResolvFileValues, error) {
	// TODO: we need to add runtime OS in case of windows.
	fInfo, err := crcos.GetFilePath("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	out, err := ioutil.ReadFile(fInfo)
	if crcos.CurrentOS() == crcos.WINDOWS {
		// TODO: we need to add logic in case of windows.
	}

	if err != nil {
		fmt.Println("This is the error here")
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
