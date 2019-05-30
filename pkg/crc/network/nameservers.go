package network

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

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

func CreateResolvFileOnInstance(driver drivers.Driver, resolvFileValues ResolvFileValues) {
	resolvFile, _ := createResolvFile(resolvFileValues)
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
