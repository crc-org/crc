package dns

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/services"
)

const (
	coreFileTemplate = `{{ .Domain }}:{{ .Port }} {
    log
    errors
    file {{ .ZonefilePath }}
}`
	zoneFileSOATemplate = `$ORIGIN .
@    3600   IN      SOA     ns.{{ .Domain }}. root.({{ .Domain }}). (
                              5         ; serial
                         604800         ; refresh
                          86400         ; retry
                        2419200         ; expire
                          36000         ; minimum
			  )
`
	zoneFileAddressTemplate = `; Name Server Information
_etcd-server-ssl._tcp.{{ .ClusterName}}.{{ .BaseDomain }}. IN SRV 10 10 2380 etcd-0.{{ .ClusterName }}.{{ .BaseDomain }}.
{{ .Hostname }}.{{ .ClusterName }}.{{ .BaseDomain }}.    IN  A     {{ .IP }}
api.{{ .ClusterName }}.{{ .BaseDomain }}.                   IN  A     {{ .IP }}
api-int.{{ .ClusterName }}.{{ .BaseDomain }}.               IN  A     {{ .IP }}
etcd-0.{{ .ClusterName }}.{{ .BaseDomain }}.                IN  A     {{ .IP }}
*.{{ .AppsDomain }}.                IN  A     {{ .IP }}
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
