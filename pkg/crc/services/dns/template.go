package dns

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/code-ready/crc/pkg/crc/constants"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/services"
	"github.com/code-ready/machine/libmachine/drivers"
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
address=/{{ .Hostname }}.{{ .ClusterName}}.{{ .BaseDomain }}/{{ .IP }}
`
)

type dnsmasqConfFileValues struct {
	BaseDomain  string
	Port        int
	ClusterName string
	Hostname    string
	IP          string
	AppsDomain  string
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
	}

	dnsConfig, err := createDnsConfigFile(dnsmasqConfFileValues)
	if err != nil {
		return err
	}

	encodeddnsConfig := base64.StdEncoding.EncodeToString([]byte(dnsConfig))
	_, err = drivers.RunSSHCommandFromDriver(serviceConfig.Driver, constants.GetPrivateKeyPath(),
		fmt.Sprintf("echo '%s' | openssl enc -base64 -d | sudo tee /var/srv/dnsmasq.conf > /dev/null",
			encodeddnsConfig))
	if err != nil {
		return err
	}
	return nil
}

func createDnsConfigFile(values dnsmasqConfFileValues) (string, error) {
	var dnsConfigFile bytes.Buffer

	t, err := template.New("dnsConfigFile").Parse(dnsmasqConfTemplate)
	if err != nil {
		return "", err
	}
	err = t.Execute(&dnsConfigFile, values)
	if err != nil {
		return "", err
	}
	return dnsConfigFile.String(), nil
}
