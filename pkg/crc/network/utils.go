package network

import (
	"fmt"
	"net"
	"net/url"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func URIStringForDisplay(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if u.User != nil {
		return fmt.Sprintf("%s://%s:xxx@%s", u.Scheme, u.User.Username(), u.Host), nil
	}
	return uri, nil
}

func matchIP(ips []net.IP, expectedIP string) bool {
	for _, ip := range ips {
		if ip.String() == expectedIP {
			return true
		}
	}

	return false
}
func CheckCRCLocalDNSReachableFromHost(apiHostname, appsHostname, appsDomain, expectedIP string) error {
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
		ip, err = net.LookupIP(appsHostname)
		if err != nil {
			// Right now admin helper fallback is not implemented on windows so
			// this check should still return an error.
			if runtime.GOOS == "windows" {
				return err
			}
			logging.Warnf("Wildcard DNS resolution for %s does not appear to be working", appsDomain)
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
