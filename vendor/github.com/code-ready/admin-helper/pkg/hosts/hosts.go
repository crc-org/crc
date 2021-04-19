package hosts

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/goodhosts/hostsfile"
)

const (
	// source https://github.com/kubernetes/apimachinery/blob/603e04655e9f537eb01238cdbce4891f832a4f27/pkg/util/validation/validation.go#L208
	dns1123SubdomainRegexp = `[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*`
	clusterDomain          = ".crc.testing"
	appsDomain             = ".apps-crc.testing"
)

var (
	clusterRegexp = regexp.MustCompile("^" + dns1123SubdomainRegexp + regexp.QuoteMeta(clusterDomain) + "$")
	appRegexp     = regexp.MustCompile("^" + dns1123SubdomainRegexp + regexp.QuoteMeta(appsDomain) + "$")
)

type Hosts struct {
	File       *hostsfile.Hosts
	HostFilter func(string) bool
}

func New() (*Hosts, error) {
	file, err := hostsfile.NewHosts()
	if err != nil {
		return nil, err
	}

	return &Hosts{
		File:       &file,
		HostFilter: defaultFilter,
	}, nil
}

func defaultFilter(s string) bool {
	return clusterRegexp.MatchString(s) || appRegexp.MatchString(s)
}

func (h *Hosts) Add(ip string, hosts []string) error {
	if err := h.verifyHosts(hosts); err != nil {
		return err
	}

	if err := h.checkIsWritable(); err != nil {
		return err
	}

	uniqueHosts := map[string]bool{}
	for i := 0; i < len(hosts); i++ {
		uniqueHosts[hosts[i]] = true
	}

	var hostEntries []string
	for key := range uniqueHosts {
		hostEntries = append(hostEntries, key)
	}

	sort.Strings(hostEntries)

	if err := h.File.Add(ip, hostEntries...); err != nil {
		return err
	}
	return h.File.Flush()
}

func (h *Hosts) Remove(hosts []string) error {
	if err := h.verifyHosts(hosts); err != nil {
		return err
	}

	if err := h.checkIsWritable(); err != nil {
		return err
	}

	uniqueHosts := map[string]bool{}
	for i := 0; i < len(hosts); i++ {
		uniqueHosts[hosts[i]] = true
	}

	var hostEntries []string
	for key := range uniqueHosts {
		hostEntries = append(hostEntries, key)
	}

	for _, host := range hostEntries {
		if err := h.File.RemoveByHostname(host); err != nil {
			return err
		}
	}
	return h.File.Flush()
}

func (h *Hosts) Clean(rawSuffixes []string) error {
	if err := h.checkIsWritable(); err != nil {
		return err
	}

	var suffixes []string
	for _, suffix := range rawSuffixes {
		if !strings.HasPrefix(suffix, ".") {
			return fmt.Errorf("suffix should start with a dot")
		}
		suffixes = append(suffixes, suffix)
	}

	var toDelete []string
	for _, line := range h.File.Lines {
		for _, host := range line.Hosts {
			for _, suffix := range suffixes {
				if strings.HasSuffix(host, suffix) {
					toDelete = append(toDelete, host)
					break
				}
			}
		}
	}

	if err := h.verifyHosts(toDelete); err != nil {
		return err
	}

	for _, host := range toDelete {
		if err := h.File.RemoveByHostname(host); err != nil {
			return err
		}
	}
	return h.File.Flush()
}

func (h *Hosts) checkIsWritable() error {
	if !h.File.IsWritable() {
		return fmt.Errorf("host file not writable, try running with elevated privileges")
	}
	return nil
}

func (h *Hosts) Contains(ip, host string) bool {
	if err := h.verifyHosts([]string{host}); err != nil {
		return false
	}

	return h.File.Has(ip, host)
}

func (h *Hosts) verifyHosts(hosts []string) error {
	for _, host := range hosts {
		if !h.HostFilter(host) {
			return fmt.Errorf("input %s rejected", host)
		}
	}
	return nil
}
