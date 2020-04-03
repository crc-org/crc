package dns

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/goodhosts"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/services"
	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	resolverFileTemplate = `port {{ .Port }}
domain {{ .Domain }}
nameserver {{ .IP }}
search_order {{ .SearchOrder }}`
)

type resolverFileValues struct {
	Port        int
	Domain      string
	IP          string
	SearchOrder int
}

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// Update /etc/hosts file for host
	if err := goodhosts.UpdateHostsFile(serviceConfig.IP, serviceConfig.BundleMetadata.GetAPIHostname(),
		serviceConfig.BundleMetadata.GetAppHostname("oauth-openshift"),
		serviceConfig.BundleMetadata.GetAppHostname("console-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("default-route-openshift-image-registry")); err != nil {
		result.Success = false
		return *result, err
	}

	// Write resolver config to host
	needRestart, err := createResolverFile(serviceConfig.IP, serviceConfig.BundleMetadata.ClusterInfo.BaseDomain,
		serviceConfig.BundleMetadata.ClusterInfo.BaseDomain)
	if err != nil {
		result.Success = false
		return *result, err
	}
	if needRestart {
		// Restart the Network on mac
		logging.Infof("Restarting the host network")
		if err := restartNetwork(); err != nil {
			result.Success = false
			return *result, fmt.Errorf("Restarting the host network failed: %v", err)
		}
		// Wait for the Network to come up but in the case of error, log it to error info.
		// If we make it as fatal call then in offline use case for mac is
		// always going to be broken.
		if err := waitForNetwork(); err != nil {
			logging.Error(err)
		}
	} else {
		logging.Infof("Network restart not needed")
	}

	// we pass the result and error on
	result.Success = true
	return *result, nil
}

func createResolverFile(InstanceIP string, domain string, filename string) (bool, error) {
	var resolverFile bytes.Buffer

	values := resolverFileValues{
		Port:        dnsServicePort,
		Domain:      domain,
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

	path := filepath.Join("/", "etc", "resolver", filename)
	return crcos.WriteFileIfContentChanged(path, resolverFile.Bytes(), 0644)
}

// restartNetwork is required to update the resolver file on OSx.
func restartNetwork() error {
	// https://medium.com/@kumar_pravin/network-restart-on-mac-os-using-shell-script-ab19ba6e6e99
	netDeviceList, _, err := crcos.RunWithDefaultLocale("networksetup", "-listallnetworkservices")
	netDeviceList = strings.TrimSpace(netDeviceList)
	if err != nil {
		return err
	}
	for _, netdevice := range strings.Split(netDeviceList, "\n")[1:] {
		time.Sleep(1 * time.Second)
		stdout, stderr, _ := crcos.RunWithDefaultLocale("networksetup", "-setnetworkserviceenabled", netdevice, "off")
		logging.Debugf("Disabling the %s Device (stdout: %s), (stderr: %s)", netdevice, stdout, stderr)
		stdout, stderr, err = crcos.RunWithDefaultLocale("networksetup", "-setnetworkserviceenabled", netdevice, "on")
		logging.Debugf("Enabling the %s Device (stdout: %s), (stderr: %s)", netdevice, stdout, stderr)
		if err != nil {
			return fmt.Errorf("%s: %v", stderr, err)
		}
	}

	return nil
}

// Wait for Network wait till the network is up, since it is required to resolve external dnsquery
func waitForNetwork() error {
	var hostResolv *network.ResolvFileValues
	var err error

	getResolvValueFromHost := func() error {
		hostResolv, err = network.GetResolvValuesFromHost()
		if err != nil {
			logging.Debugf("Not able read file: %v", err)
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	if err := errors.RetryAfter(10, getResolvValueFromHost, time.Second); err != nil {
		return fmt.Errorf("Unable to read host resolv file (%v)", err)
	}
	// retry up to 5 times
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
