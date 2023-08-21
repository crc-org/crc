package machine

import (
	"encoding/base64"
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	"github.com/crc-org/crc/v2/pkg/libmachine"
	"github.com/crc-org/machine/libmachine/drivers"
)

func getClusterConfig(bundleInfo *bundle.CrcBundleInfo) (*types.ClusterConfig, error) {
	if !bundleInfo.IsOpenShift() {
		return &types.ClusterConfig{
			ClusterType: bundleInfo.GetBundleType(),
			ProxyConfig: &httpproxy.ProxyConfig{},
		}, nil
	}

	kubeadminPassword, err := cluster.GetKubeadminPassword()
	if err != nil {
		return nil, fmt.Errorf("Error reading kubeadmin password from bundle %v", err)
	}
	proxyConfig, err := getProxyConfig(bundleInfo)
	if err != nil {
		return nil, err
	}
	clusterCACert, err := certificateAuthority(bundleInfo.GetKubeConfigPath())
	if err != nil {
		return nil, err
	}
	return &types.ClusterConfig{
		ClusterType:   bundleInfo.GetBundleType(),
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
		err := fmt.Errorf("Error getting bundle name from CRC instance, make sure you ran 'crc setup' and are using the latest bundle")
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

func getProxyConfig(bundleInfo *bundle.CrcBundleInfo) (*httpproxy.ProxyConfig, error) {
	proxy, err := httpproxy.NewProxyConfig()
	if err != nil {
		return nil, err
	}
	if proxy.IsEnabled() && bundleInfo.IsOpenShift() {
		proxy.AddNoProxy(fmt.Sprintf(".%s", bundleInfo.ClusterInfo.BaseDomain))
	}

	return proxy, nil
}
