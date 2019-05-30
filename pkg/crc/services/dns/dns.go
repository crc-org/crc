package dns

import (
	"fmt"
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
	serverType     = "dns"
	dnsServicePort = 53
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

	// run Process
	err = EnsureDNSDaemonRunning()
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

	err := createZonefileConfig(serviceConfig)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	// override resolv.conf file
	searchdomain := network.SearchDomain{Domain: serviceConfig.BundleMetadata.ClusterInfo.BaseDomain}
	nameserver := network.NameServer{IPAddress: serviceConfig.HostIP}
	resolvFileValues := network.ResolvFileValues{
		SearchDomains: []network.SearchDomain{searchdomain},
		NameServers:   []network.NameServer{nameserver}}

	network.CreateResolvFileOnInstance(serviceConfig.Driver, resolvFileValues)

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
