package dns

import (
	"context"
	"fmt"
	"time"

	"github.com/code-ready/crc/pkg/crc/adminhelper"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/services"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
)

const (
	dnsServicePort    = 53
	dnsContainerIP    = "10.88.0.8"
	publicDNSQueryURI = "quay.io"
	crcDnsmasqService = "crc-dnsmasq.service"
)

func init() {
}

func RunPostStart(serviceConfig services.ServicePostStartConfig) error {
	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		return addOpenShiftHosts(serviceConfig)
	}

	if err := setupDnsmasq(serviceConfig); err != nil {
		return err
	}

	if err := runPostStartForOS(serviceConfig); err != nil {
		return err
	}

	resolvFileValues, err := getResolvFileValues(serviceConfig)
	if err != nil {
		return err
	}
	// override resolv.conf file
	return network.CreateResolvFileOnInstance(serviceConfig.SSHRunner, resolvFileValues)
}

func setupDnsmasq(serviceConfig services.ServicePostStartConfig) error {
	if err := createDnsmasqDNSConfig(serviceConfig); err != nil {
		return err
	}
	sd := systemd.NewInstanceSystemdCommander(serviceConfig.SSHRunner)
	if state, err := sd.Status(crcDnsmasqService); err != nil || state != states.Running {
		if err := sd.Enable(crcDnsmasqService); err != nil {
			return err
		}
	}
	return sd.Start(crcDnsmasqService)
}

func getResolvFileValues(serviceConfig services.ServicePostStartConfig) (network.ResolvFileValues, error) {
	dnsServers, err := dnsServers(serviceConfig)
	if err != nil {
		return network.ResolvFileValues{}, err
	}
	return network.ResolvFileValues{
		SearchDomains: []network.SearchDomain{
			{
				Domain: fmt.Sprintf("%s.%s", serviceConfig.Name, serviceConfig.BundleMetadata.ClusterInfo.BaseDomain),
			},
		},
		NameServers: dnsServers,
	}, nil
}

func dnsServers(serviceConfig services.ServicePostStartConfig) ([]network.NameServer, error) {
	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		return []network.NameServer{
			{
				IPAddress: constants.VSockGateway,
			},
		}, nil
	}
	orgResolvValues, err := network.GetResolvValuesFromInstance(serviceConfig.SSHRunner)
	if err != nil {
		return nil, err
	}
	return append([]network.NameServer{{IPAddress: dnsContainerIP}}, orgResolvValues.NameServers...), nil
}

func CheckCRCLocalDNSReachable(ctx context.Context, serviceConfig services.ServicePostStartConfig) (string, error) {
	appsURI := fmt.Sprintf("foo.%s", serviceConfig.BundleMetadata.ClusterInfo.AppsDomain)
	// Try 30 times for 1 second interval, In nested environment most of time crc failed to get
	// Internal dns query resolved for some time.
	var queryOutput string
	var err error
	checkLocalDNSReach := func() error {
		queryOutput, _, err = serviceConfig.SSHRunner.Run(fmt.Sprintf("host -R 3 %s", appsURI))
		if err != nil {
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	if err := errors.Retry(ctx, 30*time.Second, checkLocalDNSReach, time.Second); err != nil {
		return queryOutput, err
	}
	return queryOutput, err
}

func CheckCRCPublicDNSReachable(serviceConfig services.ServicePostStartConfig) (string, error) {
	// This does not query DNS directly to account for corporate environment where external DNS resolution
	// may only be done on the host running the http(s) proxies used for internet connectivity
	proxyConfig, err := network.NewProxyConfig()
	if err != nil {
		// try without using proxy
		proxyConfig = &network.ProxyConfig{}
	}
	curlArgs := []string{"--head", publicDNSQueryURI}
	if proxyConfig.IsEnabled() {
		if proxyConfig.HTTPProxy != "" {
			curlArgs = append(curlArgs, "--proxy", proxyConfig.HTTPProxy)
		}
		if proxyConfig.HTTPSProxy != "" {
			curlArgs = append(curlArgs, "--proxy", proxyConfig.HTTPSProxy)
		}
		curlArgs = append(curlArgs, "--noproxy", proxyConfig.GetNoProxyString())
		if proxyConfig.ProxyCAFile != "" {
			curlArgs = append(curlArgs, "--proxy-cacert", proxyConfig.HTTPProxy)
		}
	}
	stdout, _, err := serviceConfig.SSHRunner.Run("curl", curlArgs...)
	return stdout, err
}

func addOpenShiftHosts(serviceConfig services.ServicePostStartConfig) error {
	return adminhelper.UpdateHostsFile(serviceConfig.IP, serviceConfig.BundleMetadata.GetAPIHostname(),
		serviceConfig.BundleMetadata.GetAppHostname("oauth-openshift"),
		serviceConfig.BundleMetadata.GetAppHostname("console-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("downloads-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("canary-openshift-ingress-canary"),
		serviceConfig.BundleMetadata.GetAppHostname("default-route-openshift-image-registry"))
}
