package dns

import (
	"fmt"
	"github.com/code-ready/machine/libmachine/drivers"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/services"

	// CoreDNS specific inmports
	"github.com/coredns/coredns/core/dnsserver"
	_ "github.com/coredns/coredns/plugin/bind"
	_ "github.com/coredns/coredns/plugin/errors"
	_ "github.com/coredns/coredns/plugin/file"
	_ "github.com/coredns/coredns/plugin/forward"
	_ "github.com/coredns/coredns/plugin/log"
	_ "github.com/coredns/coredns/plugin/template"
	"github.com/mholt/caddy"
)

const (
	serverType                  = "dns"
	dnsServicePort              = 53
	dnsConfigFilePathInInstance = "/var/srv/dnsmasq.conf"
	dnsContainerIP              = "10.88.0.8"
	dnsContainerImage           = "quay.io/crcont/dnsmasq:latest"
)

func init() {
	caddy.Quiet = true
	dnsserver.Quiet = true
	//caddy.RegisterCaddyfileLoader("flag", caddy.LoaderFunc(confLoader))

	caddy.AppName = fmt.Sprintf("CRC %s", "CoreDNS")
	caddy.AppVersion = fmt.Sprintf("1.5.0-%s", "crc_0.8.6")
}

func RunPreStart(serviceConfig services.ServicePreStartConfig) (services.ServicePreStartResult, error) {
	result := &services.ServicePreStartResult{Name: serviceConfig.Name}

	err := createCoreDNSConfig(serviceConfig)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	result.Success = true
	return *result, nil
}

func RunPostStart(serviceConfig services.ServicePostStartConfig) (services.ServicePostStartResult, error) {
	result := &services.ServicePostStartResult{Name: serviceConfig.Name}

	err := createDnsmasqDNSConfig(serviceConfig)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	// Remove the dnsmasq container if exist during the VM stop cycle
	drivers.RunSSHCommandFromDriver(serviceConfig.Driver,"sudo podman rm -f dnsmasq")

	// Start the dnsmasq container
	dnsServerRunCmd := fmt.Sprintf("sudo podman run  --ip %s --name dnsmasq -v %s:/etc/dnsmasq.conf -p 53:%d/udp --privileged -d %s",
		dnsContainerIP, dnsConfigFilePathInInstance, dnsServicePort, dnsContainerImage)
	_, err = drivers.RunSSHCommandFromDriver(serviceConfig.Driver, dnsServerRunCmd)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	orgResolvValues := network.GetResolvValuesFromInstance(serviceConfig.Driver)
	// override resolv.conf file
	searchdomain := network.SearchDomain{Domain: fmt.Sprintf("%s.%s", serviceConfig.Name, serviceConfig.BundleMetadata.ClusterInfo.BaseDomain)}
	nameserver := network.NameServer{IPAddress: dnsContainerIP}
	nameservers := []network.NameServer{nameserver}
	nameservers = append(nameservers, orgResolvValues.NameServers...)

	resolvFileValues := network.ResolvFileValues{
		SearchDomains: []network.SearchDomain{searchdomain},
		NameServers:   nameservers}

	network.CreateResolvFileOnInstance(serviceConfig.Driver, resolvFileValues)

	/*// run Process (This is for using the CoreDNS)
	// CoreDNS is not working as expected for us so in future
	// we will remove it along with all the CoreDNS part.
	err = EnsureDNSDaemonRunning()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}*/

	return runPostStartForOS(serviceConfig, result)
}

func RunDaemon() {

	caddy.DefaultConfigFile = filepath.Join(constants.CrcBaseDir, "Corefile")
	caddy.TrapSignals()

	corefile, err := caddy.LoadCaddyfile(serverType)
	if err != nil {
		logging.ErrorF("Error loading DNS configuration: %v", err)
	}

	// Start your engines
	instance, err := caddy.Start(corefile)
	if err != nil {
		logging.ErrorF("Error starting DNS service: %v", err)
	}

	instance.Wait()
}
