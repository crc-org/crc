package dns

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/services"
	"github.com/code-ready/machine/libmachine/drivers"
)

const (
	coreFileTemplate = `{{ .Domain }}:{{ .Port }} {
    log
    errors
    file {{ .ZonefilePath }}
}
.   {
        forward . /etc/resolv.conf {
             except testing
        }
    }
}`
	zoneFileSOATemplate = `$ORIGIN .
@    3600   IN      SOA     ns.{{ .Domain }}. root.{{ .Domain }}. (
                              5
                         604800
                          86400
                        2419200
                          36000
			  )
`
	zoneFileAddressTemplate = `; Name Server Information
_etcd-server-ssl._tcp.{{ .ClusterName}}.{{ .BaseDomain }}. IN SRV 10 10 2380 etcd-0.{{ .ClusterName }}.{{ .BaseDomain }}.
*.{{ .BaseDomain }}.                IN  A     {{ .IP }}
`

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

type coreFileValues struct {
	Domain       string
	Port         int
	ZonefilePath string
}

type zoneFileSOAValues struct {
	Domain   string
	Hostname string
}

type zoneFileAddressValues struct {
	BaseDomain  string
	ClusterName string
	AppsDomain  string
	Hostname    string
	IP          string
}

type dnsmasqConfFileValues struct {
	BaseDomain  string
	Port        int
	ClusterName string
	Hostname    string
	IP          string
	AppsDomain  string
}

func createCoreDNSConfig(serviceConfig services.ServicePreStartConfig) error {
	domain := serviceConfig.BundleMetadata.ClusterInfo.BaseDomain

	zoneFileSOAValues := zoneFileSOAValues{
		Domain:   domain,
		Hostname: serviceConfig.BundleMetadata.Nodes[0].Hostname,
	}

	// filepath.Join(constants.MachineBaseDir, "machines", serviceConfig.Name, "zonefile")
	zoneFilePath := filepath.Join(constants.CrcBaseDir, "zonefile")
	_, err := createInitialZoneFile(zoneFileSOAValues, zoneFilePath)
	if err != nil {
		print(err.Error())
		return err
	}

	coreFileValues := coreFileValues{
		Domain:       domain,
		Port:         dnsServicePort,
		ZonefilePath: zoneFilePath,
	}

	// filepath.Join(constants.MachineBaseDir, "machines", serviceConfig.Name, "Corefile")
	coreFilePath := filepath.Join(constants.CrcBaseDir, "Corefile")
	_, err2 := createCoreFile(coreFileValues, coreFilePath)
	if err2 != nil {
		return err2
	}

	return nil
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
	_, err = drivers.RunSSHCommandFromDriver(serviceConfig.Driver,
		fmt.Sprintf("echo '%s' | openssl enc -base64 -d | sudo tee /var/srv/dnsmasq.conf > /dev/null",
			encodeddnsConfig))
	if err != nil {
		return err
	}
	return nil
}

func createZonefileConfig(serviceConfig services.ServicePostStartConfig) error {
	zoneFileAddressValues := zoneFileAddressValues{
		ClusterName: serviceConfig.BundleMetadata.ClusterInfo.ClusterName,
		BaseDomain:  serviceConfig.BundleMetadata.ClusterInfo.BaseDomain,
		AppsDomain:  serviceConfig.BundleMetadata.ClusterInfo.AppsDomain,
		Hostname:    serviceConfig.BundleMetadata.Nodes[0].Hostname,
		IP:          serviceConfig.IP,
	}

	// filepath.Join(constants.MachineBaseDir, "machines", serviceConfig.Name, "zonefile")
	zoneFilePath := filepath.Join(constants.CrcBaseDir, "zonefile")
	_, err := addAddressesToZoneFile(zoneFileAddressValues, zoneFilePath)
	if err != nil {
		print(err.Error())
		return err
	}

	return nil
}

// refactor (duplication)
func createInitialZoneFile(values zoneFileSOAValues, path string) (bool, error) {
	var zoneFile bytes.Buffer

	t, err := template.New("zonefile").Parse(zoneFileSOATemplate)
	if err != nil {
		return false, err
	}
	t.Execute(&zoneFile, values)

	err = ioutil.WriteFile(path, zoneFile.Bytes(), 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

func addAddressesToZoneFile(values zoneFileAddressValues, path string) (bool, error) {
	var zoneFile bytes.Buffer

	t, err := template.New("zonefile").Parse(zoneFileAddressTemplate)
	if err != nil {
		return false, err
	}
	t.Execute(&zoneFile, values)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}
	if _, err := f.Write(zoneFile.Bytes()); err != nil {
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}

	return true, nil
}

func createCoreFile(values coreFileValues, path string) (bool, error) {
	var coreFile bytes.Buffer

	t, err := template.New("corefile").Parse(coreFileTemplate)
	if err != nil {
		return false, err
	}
	t.Execute(&coreFile, values)

	err = ioutil.WriteFile(path, coreFile.Bytes(), 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

func createDnsConfigFile(values dnsmasqConfFileValues) (string, error) {
	var dnsConfigFile bytes.Buffer

	t, err := template.New("dnsConfigFile").Parse(dnsmasqConfTemplate)
	if err != nil {
		return "", err
	}
	t.Execute(&dnsConfigFile, values)
	return dnsConfigFile.String(), nil
}
