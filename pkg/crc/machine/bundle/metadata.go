package bundle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/extract"
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
		OpenShiftVersion      string `json:"openshiftVersion"`
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

func getCachedBundlePath(bundleName string) string {
	path := strings.TrimSuffix(bundleName, ".crcbundle")
	return filepath.Join(constants.MachineCacheDir, path)
}

func (bundle *CrcBundleInfo) isCached() bool {
	_, err := os.Stat(bundle.cachedPath)
	return err == nil
}

func (bundle *CrcBundleInfo) readBundleInfo() error {
	bundleInfoPath := bundle.resolvePath("crc-bundle-info.json")
	f, err := ioutil.ReadFile(filepath.Clean(bundleInfoPath))
	if err != nil {
		return fmt.Errorf("Error reading %s file : %+v", bundleInfoPath, err)
	}

	err = json.Unmarshal(f, bundle)
	if err != nil {
		return fmt.Errorf("Error Unmarshal the data: %+v", err)
	}

	return nil
}

func GetCachedBundleInfo(bundleName string) (*CrcBundleInfo, error) {
	var bundleInfo CrcBundleInfo
	bundleInfo.cachedPath = getCachedBundlePath(bundleName)
	if !bundleInfo.isCached() {
		return nil, fmt.Errorf("Could not find cached bundle info")
	}
	err := bundleInfo.readBundleInfo()
	if err != nil {
		return nil, err
	}
	return &bundleInfo, nil
}

func (bundle *CrcBundleInfo) resolvePath(filename string) string {
	return filepath.Join(bundle.cachedPath, filename)
}

func Extract(sourcepath string) (*CrcBundleInfo, error) {
	err := extract.Uncompress(sourcepath, constants.MachineCacheDir)
	if err != nil {
		return nil, err
	}

	return GetCachedBundleInfo(filepath.Base(sourcepath))
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

func (bundle *CrcBundleInfo) GetBundleBuildTime() (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(bundle.BuildInfo.BuildTime))
}

func (bundle *CrcBundleInfo) GetOpenshiftVersion() string {
	return bundle.ClusterInfo.OpenShiftVersion
}
