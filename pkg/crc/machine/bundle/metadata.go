package bundle

import (
	"encoding/json"
	"fmt"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/machine/config"
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

func GetCrcBundleInfo(machineConfig config.MachineConfig) (*CrcBundleInfo, string, error) {
	var bundleInfo CrcBundleInfo
	extractedPath, err := Extract(machineConfig.BundlePath, constants.MachineCacheDir)
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
