package dns

import (
	"fmt"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/adminhelper"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
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

func RunPostStart(serviceConfig ServicePostStartConfig) error {
	if err := setupDnsmasq(serviceConfig); err != nil {
		return err
	}

	// Update /etc/hosts file for host
	if err := addOpenShiftHosts(serviceConfig); err != nil {
		return err
	}

	if err := runPostStartForOS(serviceConfig); err != nil {
		return err
	}

	return CreateResolvFileOnInstance(serviceConfig)
}

func setupDnsmasq(serviceConfig ServicePostStartConfig) error {
	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		return nil
	}

	if err := createDnsmasqDNSConfig(serviceConfig); err != nil {
		return err
	}

	// Remove the dnsmasq container if it exists during the VM stop cycle
	_, _, _ = serviceConfig.SSHRunner.Run("sudo podman rm -f dnsmasq")

	// Remove the CNI network definition forcefully
	// https://github.com/containers/libpod/issues/2767
	// TODO: We need to revisit it once podman update the CNI plugins.
	_, _, _ = serviceConfig.SSHRunner.Run(fmt.Sprintf("sudo rm -f /var/lib/cni/networks/podman/%s", dnsContainerIP))

	// Start the dnsmasq container
	dnsServerRunCmd := fmt.Sprintf("sudo podman run  --ip %s --name dnsmasq -v %s:/etc/dnsmasq.conf -p 53:%d/udp --privileged -d %s",
		dnsContainerIP, dnsConfigFilePathInInstance, dnsServicePort, dnsContainerImage)
	if _, _, err := serviceConfig.SSHRunner.Run(dnsServerRunCmd); err != nil {
		return err
	}
	return nil
}

func getResolvFileValues(serviceConfig ServicePostStartConfig) (ResolvFileValues, error) {
	dnsServers, err := dnsServers(serviceConfig)
	if err != nil {
		return ResolvFileValues{}, err
	}
	return ResolvFileValues{
		SearchDomains: []SearchDomain{
			{
				Domain: fmt.Sprintf("%s.%s", serviceConfig.Name, serviceConfig.BundleMetadata.ClusterInfo.BaseDomain),
			},
		},
		NameServers: dnsServers,
	}, nil
}

func dnsServers(serviceConfig ServicePostStartConfig) ([]NameServer, error) {
	if serviceConfig.NetworkMode == network.UserNetworkingMode {
		return []NameServer{
			{
				IPAddress: constants.VSockGateway,
			},
		}, nil
	}
	orgResolvValues, err := GetResolvValuesFromInstance(serviceConfig.SSHRunner)
	if err != nil {
		return nil, err
	}
	return append([]NameServer{{IPAddress: dnsContainerIP}}, orgResolvValues.NameServers...), nil
}

func addOpenShiftHosts(serviceConfig ServicePostStartConfig) error {
	if runtime.GOOS == "windows" && serviceConfig.NetworkMode == network.SystemNetworkingMode { // avoid UAC prompts on Windows in this mode
		return nil
	}
	return adminhelper.UpdateHostsFile(serviceConfig.IP, serviceConfig.BundleMetadata.GetAPIHostname(),
		serviceConfig.BundleMetadata.GetAppHostname("oauth-openshift"),
		serviceConfig.BundleMetadata.GetAppHostname("console-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("downloads-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("canary-openshift-ingress-canary"),
		serviceConfig.BundleMetadata.GetAppHostname("default-route-openshift-image-registry"))
}
