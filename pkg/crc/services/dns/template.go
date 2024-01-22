package dns

import (
	"bytes"
	"text/template"

	"github.com/crc-org/crc/v2/pkg/crc/services"
)

const (
	dnsmasqConfTemplate = `listen-address={{ .IP }}
expand-hosts
log-queries
local=/{{ .ClusterName}}.{{ .BaseDomain }}/
domain={{ .ClusterName}}.{{ .BaseDomain }}
address=/{{ .AppsDomain }}/{{ .IP }}
address=/api.{{ .ClusterName}}.{{ .BaseDomain }}/{{ .IP }}
address=/api-int.{{ .ClusterName}}.{{ .BaseDomain }}/{{ .IP }}
address=/{{ .Hostname }}.{{ .ClusterName}}.{{ .BaseDomain }}/{{ .InternalIP }}
`
)

type dnsmasqConfFileValues struct {
	BaseDomain  string
	Port        int
	ClusterName string
	Hostname    string
	IP          string
	AppsDomain  string
	InternalIP  string
}

func createDnsmasqDNSConfig(serviceConfig services.ServicePostStartConfig) error {
	domain := serviceConfig.BundleMetadata.ClusterInfo.BaseDomain

	dnsmasqConfFileValues := dnsmasqConfFileValues{
		BaseDomain:  domain,
		Hostname:    serviceConfig.BundleMetadata.Nodes[0].Hostname,
		AppsDomain:  serviceConfig.BundleMetadata.ClusterInfo.AppsDomain,
		ClusterName: serviceConfig.BundleMetadata.ClusterInfo.ClusterName,
		IP:          serviceConfig.IP,
		InternalIP:  serviceConfig.BundleMetadata.Nodes[0].InternalIP,
	}

	dnsConfig, err := createDNSConfigFile(dnsmasqConfFileValues, dnsmasqConfTemplate)
	if err != nil {
		return err
	}

	return serviceConfig.SSHRunner.CopyDataPrivileged([]byte(dnsConfig), "/etc/dnsmasq.d/crc-dnsmasq.conf", 0644)
}

func createDNSConfigFile(values dnsmasqConfFileValues, tmpl string) (string, error) {
	var dnsConfigFile bytes.Buffer

	t, err := template.New("dnsConfigFile").Parse(tmpl)
	if err != nil {
		return "", err
	}
	err = t.Execute(&dnsConfigFile, values)
	if err != nil {
		return "", err
	}
	return dnsConfigFile.String(), nil
}
