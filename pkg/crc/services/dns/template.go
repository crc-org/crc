package dns

import (
	"bytes"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/services"
)

const (
	dnsmasqConfTemplate = `user=root
port= {{ .Port }}
bind-interfaces
expand-hosts
log-queries
srv-host=_etcd-server-ssl._tcp.{{ .ClusterName}}.{{ .BaseDomain }},etcd-0.{{ .ClusterName}}.{{ .BaseDomain }},2380,10
local=/{{ .ClusterName}}.{{ .BaseDomain }}/
domain={{ .ClusterName}}.{{ .BaseDomain }}
address=/{{ .AppsDomain }}/{{ .IP }}
address=/etcd-0.{{ .ClusterName}}.{{ .BaseDomain }}/{{ .IP }}
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
		Port:        dnsServicePort,
		AppsDomain:  serviceConfig.BundleMetadata.ClusterInfo.AppsDomain,
		ClusterName: serviceConfig.BundleMetadata.ClusterInfo.ClusterName,
		IP:          serviceConfig.IP,
		InternalIP:  serviceConfig.BundleMetadata.Nodes[0].InternalIP,
	}

	dnsConfig, err := createDnsConfigFile(dnsmasqConfFileValues, dnsmasqConfTemplate)
	if err != nil {
		return err
	}

	return serviceConfig.SSHRunner.CopyData([]byte(dnsConfig), "/var/srv/dnsmasq.conf")
}

func createDnsConfigFile(values dnsmasqConfFileValues, tmpl string) (string, error) {
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
