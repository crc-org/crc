package dns

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/services"
)

const (
	resolverFileTemplate = `port {{ .Port }}
	nameserver {{ .IP }}
search_order {{ .SearchOrder }}`
)

type resolverFileValues struct {
	Port        int
	IP          string
	SearchOrder int
}

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// Write resolver file
	success, err := createResolverFile(serviceConfig.HostIP, filepath.Join("/", "etc", "resolver", serviceConfig.BundleMetadata.ClusterInfo.BaseDomain))
	// we pass the result and error on
	result.Success = success
	return *result, err
}

func createResolverFile(hostIP string, path string) (bool, error) {
	var resolverFile bytes.Buffer

	values := resolverFileValues{
		Port:        dnsServicePort,
		IP:          hostIP,
		SearchOrder: 1,
	}

	t, err := template.New("resolver").Parse(resolverFileTemplate)
	if err != nil {
		return false, err
	}
	t.Execute(&resolverFile, values)

	err = ioutil.WriteFile(path, resolverFile.Bytes(), 0644)

	return true, nil
}
