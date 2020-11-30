package bundle

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Metadata structure to unmarshal the crc-bundle-info.json file

type CrcBundleInfo struct {
	Version     string      `json:"version"`
	Type        string      `json:"type"`
	BuildInfo   BuildInfo   `json:"buildInfo"`
	ClusterInfo ClusterInfo `json:"clusterInfo"`
	Nodes       []Node      `json:"nodes"`
	Storage     Storage     `json:"storage"`
	DriverInfo  DriverInfo  `json:"driverInfo"`
	cachedPath  string
}

type BuildInfo struct {
	BuildTime                 string `json:"buildTime"`
	OpenshiftInstallerVersion string `json:"openshiftInstallerVersion"`
	SncVersion                string `json:"sncVersion"`
}

type ClusterInfo struct {
	OpenShiftVersion      string `json:"openshiftVersion"`
	ClusterName           string `json:"clusterName"`
	BaseDomain            string `json:"baseDomain"`
	AppsDomain            string `json:"appsDomain"`
	SSHPrivateKeyFile     string `json:"sshPrivateKeyFile"`
	KubeConfig            string `json:"kubeConfig"`
	KubeadminPasswordFile string `json:"kubeadminPasswordFile"`
	OpenshiftPullSecret   string `json:"openshiftPullSecret,omitempty"`
}

type Node struct {
	Kind          []string `json:"kind"`
	Hostname      string   `json:"hostname"`
	DiskImage     string   `json:"diskImage"`
	KernelCmdLine string   `json:"kernelCmdLine,omitempty"`
	Initramfs     string   `json:"initramfs,omitempty"`
	Kernel        string   `json:"kernel,omitempty"`
	InternalIP    string   `json:"internalIP"`
}

type Storage struct {
	DiskImages []DiskImage `json:"diskImages"`
}

type DiskImage struct {
	Name     string `json:"name"`
	Format   string `json:"format"`
	Size     string `json:"size"`
	Checksum string `json:"sha256sum"`
}

type DriverInfo struct {
	Name string `json:"name"`
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

func (bundle *CrcBundleInfo) GetInternalIP() string {
	if bundle.Nodes[0].InternalIP == "" {
		return ""
	}
	return bundle.resolvePath(bundle.Nodes[0].InternalIP)
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

func (bundle *CrcBundleInfo) GetDiskSize() (int64, error) {
	size, err := strconv.ParseInt(bundle.Storage.DiskImages[0].Size, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (bundle *CrcBundleInfo) CheckDiskImageSize() error {
	diskImagePath := bundle.GetDiskImagePath()
	expectedSize, err := bundle.GetDiskSize()
	if err != nil {
		return err
	}
	f, err := os.Stat(diskImagePath)
	if err != nil {
		return err
	}
	gotSize := f.Size()
	if expectedSize != gotSize {
		return fmt.Errorf("Expected size %d Got %d", expectedSize, gotSize)
	}
	return nil
}
