package dns

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/services"
	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	dnsServicePort              = 53
	dnsConfigFilePathInInstance = "/var/srv/dnsmasq.conf"
	dnsContainerIP              = "10.88.0.8"
	dnsContainerImage           = "quay.io/crcont/dnsmasq:latest"
	publicDNSQueryURI           = "quay.io"
)

func init() {
}

func RunPostStart(serviceConfig services.ServicePostStartConfig) (services.ServicePostStartResult, error) {
	result := &services.ServicePostStartResult{Name: serviceConfig.Name}

	err := createDnsmasqDNSConfig(serviceConfig)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	// Remove the dnsmasq container if it exists during the VM stop cycle
	_, _ = serviceConfig.SSHRunner.Run("sudo podman rm -f dnsmasq")

	// Remove the CNI network definition forcefully
	// https://github.com/containers/libpod/issues/2767
	// TODO: We need to revisit it once podman update the CNI plugins.
	_, _ = serviceConfig.SSHRunner.Run(fmt.Sprintf("sudo rm -f /var/lib/cni/networks/podman/%s", dnsContainerIP))

	// Start the dnsmasq container
	dnsServerRunCmd := fmt.Sprintf("sudo podman run  --ip %s --name dnsmasq -v %s:/etc/dnsmasq.conf -p 53:%d/udp --privileged -d %s",
		dnsContainerIP, dnsConfigFilePathInInstance, dnsServicePort, dnsContainerImage)
	_, err = serviceConfig.SSHRunner.Run(dnsServerRunCmd)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	// We need to restart the Host Network before updating
	// the VM's /etc/resolv.conf file.
	res, err := runPostStartForOS(serviceConfig, result)
	if err != nil {
		result.Success = res.Success
		result.Error = err.Error()
		return *result, err
	}

	orgResolvValues, err := network.GetResolvValuesFromInstance(serviceConfig.SSHRunner)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}
	// override resolv.conf file
	searchdomain := network.SearchDomain{Domain: fmt.Sprintf("%s.%s", serviceConfig.Name, serviceConfig.BundleMetadata.ClusterInfo.BaseDomain)}
	nameserver := network.NameServer{IPAddress: dnsContainerIP}
	nameservers := []network.NameServer{nameserver}
	nameservers = append(nameservers, orgResolvValues.NameServers...)

	resolvFileValues := network.ResolvFileValues{
		SearchDomains: []network.SearchDomain{searchdomain},
		NameServers:   nameservers}

	if err := network.CreateResolvFileOnInstance(serviceConfig.SSHRunner, resolvFileValues); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	result.Success = true
	return *result, nil
}

func CheckCRCLocalDNSReachable(serviceConfig services.ServicePostStartConfig) (string, error) {
	appsURI := fmt.Sprintf("foo.%s", serviceConfig.BundleMetadata.ClusterInfo.AppsDomain)
	// Try 30 times for 1 second interval, In nested environment most of time crc failed to get
	// Internal dns query resolved for some time.
	var queryOutput string
	var err error
	checkLocalDNSReach := func() error {
		queryOutput, err = serviceConfig.SSHRunner.Run(fmt.Sprintf("host -R 3 %s", appsURI))
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	if err := errors.RetryAfter(30*time.Second, checkLocalDNSReach, time.Second); err != nil {
		return queryOutput, err
	}
	return queryOutput, err
}

func CheckCRCPublicDNSReachable(serviceConfig services.ServicePostStartConfig) (string, error) {
	return serviceConfig.SSHRunner.Run(fmt.Sprintf("host -R 3 %s", publicDNSQueryURI))
}

func matchIP(ips []net.IP, expectedIP string) bool {
	for _, ip := range ips {
		if ip.String() == expectedIP {
			return true
		}
	}
	return false
}

func CheckCRCLocalDNSReachableFromHost(bundle *bundle.CrcBundleInfo, expectedIP string) error {
	apiHostname := bundle.GetAPIHostname()
	ip, err := net.LookupIP(apiHostname)
	if err != nil {
		return err
	}
	logging.Debugf("%s resolved to %s", apiHostname, ip)
	if !matchIP(ip, expectedIP) {
		logging.Warnf("%s resolved to %s but %s was expected", apiHostname, ip, expectedIP)
		return fmt.Errorf("Invalid IP for %s", apiHostname)
	}

	if runtime.GOOS != "darwin" {
		/* This check will fail with !CGO_ENABLED builds on darwin as
		 * in this case, /etc/resolver/ will not be used, so we won't
		 * have wildcard DNS for our domains
		 */
		appsHostname := bundle.GetAppHostname("foo")
		ip, err = net.LookupIP(appsHostname)
		if err != nil {
			// Right now goodhosts fallback is not implemented in windows so
			// this checks should still return the error.
			if crcos.CurrentOS() == crcos.WINDOWS {
				return err
			}
			logging.Warnf("Wildcard DNS resolution for %s does not appear to be working", bundle.ClusterInfo.AppsDomain)
			return nil
		}
		logging.Debugf("%s resolved to %s", appsHostname, ip)
		if !matchIP(ip, expectedIP) {
			logging.Warnf("%s resolved to %s but %s was expected", appsHostname, ip, expectedIP)
			return fmt.Errorf("Invalid IP for %s", appsHostname)
		}
	}
	return nil
}
