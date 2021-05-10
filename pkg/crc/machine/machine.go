package machine

import (
	"encoding/base64"
	"fmt"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/libmachine"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/code-ready/machine/libmachine/drivers"
)

func getClusterConfig(bundleInfo *bundle.CrcBundleInfo) (*types.ClusterConfig, error) {
	kubeadminPassword, err := cluster.GetKubeadminPassword()
	if err != nil {
		return nil, fmt.Errorf("Error reading kubeadmin password from bundle %v", err)
	}
	proxyConfig, err := getProxyConfig(bundleInfo.ClusterInfo.BaseDomain)
	if err != nil {
		return nil, err
	}
	clusterCACert, err := certificateAuthority(bundleInfo.GetKubeConfigPath())
	if err != nil {
		return nil, err
	}
	return &types.ClusterConfig{
		ClusterCACert: base64.StdEncoding.EncodeToString(clusterCACert),
		KubeConfig:    bundleInfo.GetKubeConfigPath(),
		KubeAdminPass: kubeadminPassword,
		WebConsoleURL: fmt.Sprintf("https://%s", bundleInfo.GetAppHostname("console-openshift-console")),
		ClusterAPI:    fmt.Sprintf("https://%s:6443", bundleInfo.GetAPIHostname()),
		ProxyConfig:   proxyConfig,
	}, nil
}

func getBundleMetadataFromDriver(driver drivers.Driver) (*bundle.CrcBundleInfo, error) {
	bundleName, err := driver.GetBundleName()
	if err != nil {
		err := fmt.Errorf("Error getting bundle name from CodeReady Containers instance, make sure you ran 'crc setup' and are using the latest bundle")
		return nil, err
	}
	metadata, err := bundle.Get(bundleName)
	if err != nil {
		return nil, err
	}

	return metadata, err
}

func createLibMachineClient() (libmachine.API, func()) {
	client := libmachine.NewClient(constants.MachineBaseDir)
	return client, func() {
		client.Close()
	}
}

func getSSHPort(vsockNetwork bool) int {
	if vsockNetwork {
		return constants.VsockSSHPort
	}
	return constants.DefaultSSHPort
}

func getIP(h *host.Host, vsockNetwork bool) (string, error) {
	if vsockNetwork {
		return "127.0.0.1", nil
	}
	return h.Driver.GetIP()
}

func getProxyConfig(baseDomainName string) (*network.ProxyConfig, error) {
	proxy, err := network.NewProxyConfig()
	if err != nil {
		return nil, err
	}
	if proxy.IsEnabled() {
		proxy.AddNoProxy(fmt.Sprintf(".%s", baseDomainName))
	}

	return proxy, nil
}
