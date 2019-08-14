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
}

func GetCrcBundleInfo(bundlePath string) (*CrcBundleInfo, string, error) {
	var bundleInfo CrcBundleInfo
	extractedPath, err := Extract(bundlePath, constants.MachineCacheDir)
	if err != nil {
		return nil, extractedPath, fmt.Errorf("Error during extraction : %+v", err)
	}
	BundleInfoPath := filepath.Join(extractedPath, "crc-bundle-info.json")
	f, err := ioutil.ReadFile(BundleInfoPath)
	if err != nil {
		return nil, extractedPath, fmt.Errorf("Error reading %s file : %+v", BundleInfoPath, err)
	}

	err = json.Unmarshal(f, &bundleInfo)
	if err != nil {
		return nil, extractedPath, fmt.Errorf("Error Unmarshal the data: %+v", err)
	}
	return &bundleInfo, extractedPath, nil
}

func (bundle *CrcBundleInfo) GetAPIHostname() string {
	return fmt.Sprintf("api.%s.%s", bundle.ClusterInfo.ClusterName, bundle.ClusterInfo.BaseDomain)
}

func (bundle *CrcBundleInfo) GetAppHostname(appName string) string {
	return fmt.Sprintf("%s.%s", appName, bundle.ClusterInfo.AppsDomain)
}
