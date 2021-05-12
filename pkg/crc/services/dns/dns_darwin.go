package dns

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
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

func runPostStartForOS(serviceConfig ServicePostStartConfig) error {
	// Update /etc/hosts file for host
	if err := addOpenShiftHosts(serviceConfig); err != nil {
		return err
	}

	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		return nil
	}

	// Write resolver config to host
	needRestart, err := createResolverFile(serviceConfig.IP, serviceConfig.BundleMetadata.ClusterInfo.BaseDomain,
		serviceConfig.BundleMetadata.ClusterInfo.BaseDomain)
	if err != nil {
		return err
	}
	if needRestart {
		// Restart the Network on mac
		logging.Infof("Restarting the host network")
		if err := restartNetwork(); err != nil {
			return errors.Wrap(err, "Restarting the host network failed")
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

	return nil
}

func createResolverFile(instanceIP string, domain string, filename string) (bool, error) {
	var resolverFile bytes.Buffer

	values := resolverFileValues{
		Port:        dnsServicePort,
		Domain:      domain,
		IP:          instanceIP,
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

func checkNetworkConnectivity() error {
	hostResolv, err := GetResolvValuesFromHost()
	if err != nil {
		logging.Debugf("Unable to read resolv.conf: %v", err)
		return err
	}

	for _, ns := range hostResolv.NameServers {
		_, _, err := crcos.RunWithDefaultLocale("ping", "-c4", ns.IPAddress)
		if err != nil {
			continue
		}
		logging.Debugf("Successfully pinged %s, network is up", ns.IPAddress)
		return nil
	}
	return fmt.Errorf("Could not ping any nameservers")
}

// Wait for Network wait till the network is up, since it is required to resolve external dnsquery
func waitForNetwork() error {
	retriableConnectivityCheck := func() error {
		err := checkNetworkConnectivity()
		if err != nil {
			return &crcerrors.RetriableError{Err: err}
		}
		return nil
	}
	if err := crcerrors.RetryAfter(15*time.Second, retriableConnectivityCheck, time.Second); err != nil {
		return fmt.Errorf("Host is not connected to internet")
	}

	return nil
}
