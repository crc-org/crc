package dns

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
)

func CheckCRCLocalDNSReachable(serviceConfig ServicePostStartConfig) (string, error) {
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

	if err := errors.RetryAfter(30*time.Second, checkLocalDNSReach, time.Second); err != nil {
		return queryOutput, err
	}
	return queryOutput, err
}

func CheckCRCPublicDNSReachable(serviceConfig ServicePostStartConfig) (string, error) {
	stdout, _, err := serviceConfig.SSHRunner.Run(fmt.Sprintf("host -R 3 %s", publicDNSQueryURI))
	return stdout, err
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
			// Right now admin helper fallback is not implemented on windows so
			// this check should still return an error.
			if runtime.GOOS == "windows" {
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

func matchIP(ips []net.IP, expectedIP string) bool {
	for _, ip := range ips {
		if ip.String() == expectedIP {
			return true
		}
	}

	return false
}
