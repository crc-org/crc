package bundle

import (
	"encoding/json"
	"fmt"
	"github.com/code-ready/crc/pkg/crc/constants"
	"io/ioutil"
	"path/filepath"
)

// Metadata structure to unmarshal the crc-bundle-info.json file

type CrcBundleInfo struct {
	Version   string `json:"version"`
	Type      string `json:"type"`
	BuildInfo struct {
		BuildTime                 string `json:"buildTime"`
		OpenshiftInstallerVersion string `json:"openshiftInstallerVersion"`
		SncVersion                string `json:"sncVersion"`
	} `json:"buildInfo"`
	ClusterInfo struct {
		ClusterName           string `json:"clusterName"`
		BaseDomain            string `json:"baseDomain"`
		AppsDomain            string `json:"appsDomain"`
		SSHPrivateKeyFile     string `json:"sshPrivateKeyFile"`
		KubeConfig            string `json:"kubeConfig"`
		KubeadminPasswordFile string `json:"kubeadminPasswordFile"`
	} `json:"clusterInfo"`
	Nodes []struct {
		Kind          []string `json:"kind"`
		Hostname      string   `json:"hostname"`
		DiskImage     string   `json:"diskImage"`
		KernelCmdLine string   `json:"kernelCmdLine"`
		Initramfs     string   `json:"initramfs"`
		Kernel        string   `json:"kernel"`
	} `json:"nodes"`
	Storage struct {
		DiskImages []struct {
			Name   string `json:"name"`
			Format string `json:"format"`
		} `json:"diskImages"`
	} `json:"storage"`
	cachedPath string
}

func GetCrcBundleInfo(bundlePath string) (*CrcBundleInfo, error) {
	var bundleInfo CrcBundleInfo
	extractedPath, err := Extract(bundlePath, constants.MachineCacheDir)
	if err != nil {
		return nil, fmt.Errorf("Error during extraction : %+v", err)
	}
	BundleInfoPath := filepath.Join(extractedPath, "crc-bundle-info.json")
	f, err := ioutil.ReadFile(BundleInfoPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s file : %+v", BundleInfoPath, err)
	}

	err = json.Unmarshal(f, &bundleInfo)
	if err != nil {
		return nil, fmt.Errorf("Error Unmarshal the data: %+v", err)
	}
	bundleInfo.cachedPath = extractedPath
	return &bundleInfo, nil
}

func (bundle *CrcBundleInfo) resolvePath(filename string) string {
	return filepath.Join(bundle.cachedPath, filename)
}

func (bundle *CrcBundleInfo) GetAPIHostname() string {
	return fmt.Sprintf("api.%s.%s", bundle.ClusterInfo.ClusterName, bundle.ClusterInfo.BaseDomain)
}

func (bundle *CrcBundleInfo) GetAppHostname(appName string) string {
	return fmt.Sprintf("%s.%s", appName, bundle.ClusterInfo.AppsDomain)
}

func (bundle *CrcBundleInfo) GetDiskImagePath() string {
	return bundle.resolvePath(bundle.Storage.DiskImages[0].Name)
}

func (bundle *CrcBundleInfo) GetKubeConfigPath() string {
	return bundle.resolvePath(bundle.ClusterInfo.KubeConfig)
}

func (bundle *CrcBundleInfo) GetSSHKeyPath() string {
	return bundle.resolvePath(bundle.ClusterInfo.SSHPrivateKeyFile)
}

func (bundle *CrcBundleInfo) GetKernelPath() string {
	if bundle.Nodes[0].Kernel == "" {
		return ""
	}
	return bundle.resolvePath(bundle.Nodes[0].Kernel)
}

func (bundle *CrcBundleInfo) GetInitramfsPath() string {
	if bundle.Nodes[0].Initramfs == "" {
		return ""
	}
	return bundle.resolvePath(bundle.Nodes[0].Initramfs)
}

func (bundle *CrcBundleInfo) GetKubeadminPassword() (string, error) {
	rawData, err := ioutil.ReadFile(bundle.resolvePath(bundle.ClusterInfo.KubeadminPasswordFile))
	return string(rawData), err
}
