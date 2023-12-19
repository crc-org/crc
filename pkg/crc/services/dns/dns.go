package dns

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/adminhelper"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	"github.com/crc-org/crc/v2/pkg/crc/services"
	"github.com/crc-org/crc/v2/pkg/crc/systemd"
	"github.com/crc-org/crc/v2/pkg/crc/systemd/states"
)

const (
	dnsServicePort    = 53
	publicDNSQueryURI = "quay.io"
	dnsmasqService    = "dnsmasq.service"
)

func init() {
}

func RunPostStart(serviceConfig services.ServicePostStartConfig) error {
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
	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		return nil
	}

	if err := createDnsmasqDNSConfig(serviceConfig); err != nil {
		return err
	}
	sd := systemd.NewInstanceSystemdCommander(serviceConfig.SSHRunner)
	if state, err := sd.Status(dnsmasqService); err != nil || state != states.Running {
		if err := sd.Enable(dnsmasqService); err != nil {
			return err
		}
	}
	return sd.Start(dnsmasqService)
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
	return append([]network.NameServer{{IPAddress: serviceConfig.IP}}, orgResolvValues.NameServers...), nil
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
	proxyConfig, err := httpproxy.NewProxyConfig()
	if err != nil {
		// try without using proxy
		proxyConfig = &httpproxy.ProxyConfig{}
	}
	curlArgs := []string{"--head", publicDNSQueryURI}
	if proxyConfig.IsEnabled() {
		proxyHost := proxyConfig.HTTPProxy
		if proxyConfig.HTTPSProxy != "" {
			proxyHost = proxyConfig.HTTPSProxy
		}
		if proxyHost != "" {
			curlArgs = append(curlArgs, "--proxy", proxyHost)
		}
		curlArgs = append(curlArgs, "--noproxy", proxyConfig.GetNoProxyString())
		if proxyConfig.ProxyCAFile != "" {
			// --proxy-cacert/--cacert replaces the system CAs with the specified one.
			// If not using MITM proxy, --cacert must *not* be used, and if not using
			// https:// proxy, --proxy-cacert must *not* be used
			// ProxyCAFile is ambiguous, we cannot know if it's set because of MITM proxy,
			// because of https:// proxy, or because of both
			// We do not really care about transport security for this test, all that
			// matters is whether or not we can resolve the hostname, so we can
			// workaround this ambiguity by using an insecure connection
			curlArgs = append(curlArgs, "--insecure", "--proxy-insecure")
		}
	}
	stdout, _, err := serviceConfig.SSHRunner.Run("curl", curlArgs...)
	return stdout, err
}

func CheckCRCLocalDNSReachableFromHost(serviceConfig services.ServicePostStartConfig) error {
	bundle := serviceConfig.BundleMetadata
	apiHostname := bundle.GetAPIHostname()
	ip, err := net.LookupIP(apiHostname)
	if err != nil {
		return err
	}
	logging.Debugf("%s resolved to %s", apiHostname, ip)
	if !matchIP(ip, serviceConfig.IP) {
		logging.Warnf("%s resolved to %s but %s was expected", apiHostname, ip, serviceConfig.IP)
		return fmt.Errorf("Invalid IP for %s", apiHostname)
	}

	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		// user-mode networking does not setup wildcard DNS on the host. It relies on admin-helper
		// to create entries in /etc/hosts for routes defined in the cluster.
		return nil
	}

	if runtime.GOOS == "darwin" {
		/* This check will fail with !CGO_ENABLED builds on darwin as
		 * in this case, /etc/resolver/ will not be used, so we won't
		 * have wildcard DNS for our domains
		 */
		/* This can be removed when we switch to go 1.20
		 *  https://github.com/golang/go/issues/12524
		 */
		return nil
	}

	appsHostname := bundle.GetAppHostname("foo")
	ip, err = net.LookupIP(appsHostname)
	if err != nil {
		// Right now admin helper fallback is not implemented on windows so
		// this check should still return an error.
		if runtime.GOOS == "windows" {
			return err
		}
		logging.Warnf("Wildcard DNS resolution for %s does not appear to be working", bundle.ClusterInfo.AppsDomain)
		return nil
	}
	logging.Debugf("%s resolved to %s", appsHostname, ip)
	if !matchIP(ip, serviceConfig.IP) {
		logging.Warnf("%s resolved to %s but %s was expected", appsHostname, ip, serviceConfig.IP)
		return fmt.Errorf("Invalid IP for %s", appsHostname)
	}
	return nil
}

func matchIP(ips []net.IP, expectedIP string) bool {
	for _, ip := range ips {
		if ip.String() == expectedIP {
			return true
		}
	}

	return false
}

func addOpenShiftHosts(serviceConfig services.ServicePostStartConfig) error {
	return adminhelper.UpdateHostsFile(serviceConfig.IP, serviceConfig.BundleMetadata.GetAPIHostname(),
		serviceConfig.BundleMetadata.GetAppHostname("oauth-openshift"),
		serviceConfig.BundleMetadata.GetAppHostname("console-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("downloads-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("canary-openshift-ingress-canary"),
		serviceConfig.BundleMetadata.GetAppHostname("default-route-openshift-image-registry"))
}

func AddPodmanHosts(ip string) error {
	return adminhelper.UpdateHostsFile(ip, "podman.crc.testing")
}
