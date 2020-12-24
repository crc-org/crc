package machine

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/libmachine"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/log"
)

func getClusterConfig(bundleInfo *bundle.CrcBundleInfo) (*ClusterConfig, error) {
	kubeadminPassword, err := bundleInfo.GetKubeadminPassword()
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
	return &ClusterConfig{
		ClusterCACert: base64.StdEncoding.EncodeToString(clusterCACert),
		KubeConfig:    bundleInfo.GetKubeConfigPath(),
		KubeAdminPass: kubeadminPassword,
		WebConsoleURL: constants.DefaultWebConsoleURL,
		ClusterAPI:    constants.DefaultAPIURL,
		ProxyConfig:   proxyConfig,
	}, nil
}

func getBundleMetadataFromDriver(driver drivers.Driver) (string, *bundle.CrcBundleInfo, error) {
	bundleName, err := driver.GetBundleName()
	if err != nil {
		err := fmt.Errorf("Error getting bundle name from CodeReady Containers instance, make sure you ran 'crc setup' and are using the latest bundle")
		return "", nil, err
	}
	metadata, err := bundle.GetCachedBundleInfo(bundleName)
	if err != nil {
		return "", nil, err
	}

	return bundleName, metadata, err
}

func createLibMachineClient(debug bool) (*libmachine.Client, func(), error) {
	logfile, err := setMachineLogging(debug)
	if err != nil {
		return nil, func() {}, err
	}
	client := libmachine.NewClient(constants.MachineBaseDir)
	return client, func() {
		client.Close()
		if logfile != nil {
			logfile.Close()
		}
	}, nil
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

func setMachineLogging(logs bool) (*os.File, error) {
	if !logs {
		log.SetDebug(true)
		logfile, err := logging.OpenLogFile(constants.LogFilePath)
		if err != nil {
			return nil, err
		}
		log.SetOutWriter(logfile)
		log.SetErrWriter(logfile)
		return logfile, nil
	}
	log.SetDebug(true)
	return nil, nil
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
