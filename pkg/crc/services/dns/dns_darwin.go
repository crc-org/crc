package dns

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/services"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/code-ready/goodhosts"
)

const (
	resolverFileTemplate = `port {{ .Port }}
nameserver {{ .IP }}
search_order {{ .SearchOrder }}`
)

type resolverFileValues struct {
	Port        int
	IP          string
	SearchOrder int
}

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// Update /etc/hosts file for host
	if err := updateHostsConfFile(serviceConfig.IP, serviceConfig.BundleMetadata.GetAPIHostname(),
		serviceConfig.BundleMetadata.GetAppHostname("oauth-openshift")); err != nil {
		result.Success = false
		return *result, err
	}

	// Write resolver config to host
	needRestart, err := createResolverFile(serviceConfig.IP, filepath.Join("/", "etc", "resolver", serviceConfig.BundleMetadata.ClusterInfo.BaseDomain))
	if err != nil {
		result.Success = false
		return *result, err
	}
	if needRestart {
		// Restart the Network on mac
		logging.InfoF("Restarting the host network")
		if err := restartNetwork(); err != nil {
			result.Success = false
			return *result, err
		}
	} else {
		logging.InfoF("Network restart not needed")
	}

	// we pass the result and error on
	result.Success = true
	return *result, nil
}

func createResolverFile(InstanceIP string, path string) (bool, error) {
	var resolverFile bytes.Buffer

	values := resolverFileValues{
		Port:        dnsServicePort,
		IP:          InstanceIP,
		SearchOrder: 1,
	}

	t, err := template.New("resolver").Parse(resolverFileTemplate)
	if err != nil {
		return false, err
	}
	err = t.Execute(&resolverFile, values)
	if err != nil {
		return false, err
	}

	return crcos.WriteFileIfContentChanged(path, resolverFile.Bytes(), 0644)
}

func restartNetwork() error {
	// https://medium.com/@kumar_pravin/network-restart-on-mac-os-using-shell-script-ab19ba6e6e99
	netDeviceList, _, err := crcos.RunWithDefaultLocale("networksetup", "-listallnetworkservices")
	netDeviceList = strings.TrimSpace(netDeviceList)
	if err != nil {
		return err
	}
	for _, netdevice := range strings.Split(netDeviceList, "\n")[1:] {
		time.Sleep(1 * time.Second)
		stdout, stderr, err := crcos.RunWithDefaultLocale("networksetup", "-setnetworkserviceenabled", netdevice, "off")
		logging.DebugF("Disabling the %s Device (stdout: %s), (stderr: %s)", netdevice, stdout, stderr)
		stdout, stderr, err = crcos.RunWithDefaultLocale("networksetup", "-setnetworkserviceenabled", netdevice, "on")
		logging.DebugF("Enabling the %s Device (stdout: %s), (stderr: %s)", netdevice, stdout, stderr)
		if err != nil {
			return fmt.Errorf("%s: %v", stderr, err)
		}
	}

	return waitForNetwork()
}

// Wait for Network wait till the network is up, since it is required to resolve external dnsquery
func waitForNetwork() error {
	hostResolv, err := network.GetResolvValuesFromHost()
	if err != nil {
		// wait and retry
		return fmt.Errorf("Unable to read host resolv file (%v)", err)
	}
	// retry for 5 times
	for i := 0; i < 5; i++ {
		for _, ns := range hostResolv.NameServers {
			_, _, err := crcos.RunWithDefaultLocale("ping", "-c4", ns.IPAddress)
			if err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("Host is not connected to internet.")
}

// updateHostsConfFile updates the host's /etc/hosts file with Instance IP.
func updateHostsConfFile(instanceIP string, hostnames ...string) error {
	// Get the current value of /etc/hosts file
	hosts, err := goodhosts.NewHosts()
	if err != nil {
		return err
	}
	for _, hostname := range hostnames {
		if err := replaceHostname(&hosts, instanceIP, hostname); err != nil {
			return err
		}
	}

	// Flush operation is required just after replacing the host this
	// add the entry to /etc/hosts file
	if err := hosts.Flush(); err != nil {
		return err
	}
	return nil
}

func replaceHostname(hosts *goodhosts.Hosts, ipAddress string, hostname string) error {
	logging.DebugF("Updating %s to %s in /etc/hosts file", hostname, ipAddress)
	if hosts.HasHostname(hostname) {
		if err := hosts.RemoveByHostname(hostname); err != nil {
			return err
		}
	}

	if err := hosts.Add(ipAddress, hostname); err != nil {
		return err
	}

	return nil
}
