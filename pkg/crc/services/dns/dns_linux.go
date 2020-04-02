package dns

import (
	"github.com/code-ready/crc/pkg/crc/goodhosts"
	"github.com/code-ready/crc/pkg/crc/services"
)

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// We might need to set the firewall here to forward
	// Update /etc/hosts file for host
	if err := goodhosts.UpdateHostsFile(serviceConfig.IP, serviceConfig.BundleMetadata.GetAPIHostname(),
		serviceConfig.BundleMetadata.GetAppHostname("oauth-openshift"),
		serviceConfig.BundleMetadata.GetAppHostname("console-openshift-console"),
		serviceConfig.BundleMetadata.GetAppHostname("default-route-openshift-image-registry")); err != nil {
		result.Success = false
		return *result, err
	}
	result.Success = true
	return *result, nil
}
