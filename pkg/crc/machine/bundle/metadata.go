package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/download"
)

type Type string

const (
	OpenShift Type = "snc"
	Podman    Type = "podman"
)

// Metadata structure to unmarshal the crc-bundle-info.json file

type CrcBundleInfo struct {
	Version     string      `json:"version"`
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	BuildInfo   BuildInfo   `json:"buildInfo"`
	ClusterInfo ClusterInfo `json:"clusterInfo"`
	Nodes       []Node      `json:"nodes"`
	Storage     Storage     `json:"storage"`
	DriverInfo  DriverInfo  `json:"driverInfo"`

	cachedPath string
}

type BuildInfo struct {
	BuildTime                 string `json:"buildTime"`
	OpenshiftInstallerVersion string `json:"openshiftInstallerVersion"`
	SncVersion                string `json:"sncVersion"`
}

type ClusterInfo struct {
	OpenShiftVersion    *semver.Version `json:"openshiftVersion"`
	ClusterName         string          `json:"clusterName"`
	BaseDomain          string          `json:"baseDomain"`
	AppsDomain          string          `json:"appsDomain"`
	SSHPrivateKeyFile   string          `json:"sshPrivateKeyFile"`
	KubeConfig          string          `json:"kubeConfig"`
	OpenshiftPullSecret string          `json:"openshiftPullSecret,omitempty"`
}

type Node struct {
	Kind          []string `json:"kind"`
	Hostname      string   `json:"hostname"`
	DiskImage     string   `json:"diskImage"`
	KernelCmdLine string   `json:"kernelCmdLine,omitempty"`
	Initramfs     string   `json:"initramfs,omitempty"`
	Kernel        string   `json:"kernel,omitempty"`
	InternalIP    string   `json:"internalIP"`
	PodmanVersion string   `json:"podmanVersion,omitempty"`
}

type Storage struct {
	DiskImages []DiskImage    `json:"diskImages"`
	Files      []FileListItem `json:"fileList"`
}

type File struct {
	Name     string `json:"name"`
	Size     string `json:"size"`
	Checksum string `json:"sha256sum"`
}

type DiskImage struct {
	File
	Format string `json:"format"`
}

type FileType string

const (
	OcExecutable     FileType = "oc-executable"
	PodmanExecutable FileType = "podman-executable"
)

type FileListItem struct {
	File
	Type FileType `json:"type"`
}

type DriverInfo struct {
	Name string `json:"name"`
}

func (bundle *CrcBundleInfo) resolvePath(filename string) string {
	return filepath.Join(bundle.cachedPath, filename)
}

func (bundle *CrcBundleInfo) GetBundleName() string {
	return bundle.Name
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

func (bundle *CrcBundleInfo) GetDiskImageFormat() string {
	return bundle.Storage.DiskImages[0].Format
}

func (bundle *CrcBundleInfo) GetKubeConfigPath() string {
	return bundle.resolvePath(bundle.ClusterInfo.KubeConfig)
}

func (bundle *CrcBundleInfo) GetOcPath() string {
	for _, file := range bundle.Storage.Files {
		if file.Type == OcExecutable {
			return bundle.resolvePath(file.Name)
		}
	}

	return ""
}

func (bundle *CrcBundleInfo) GetPodmanPath() string {
	for _, file := range bundle.Storage.Files {
		if file.Type == PodmanExecutable {
			return bundle.resolvePath(file.Name)
		}
	}

	return ""
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

func (bundle *CrcBundleInfo) GetKernelCommandLine() string {
	return bundle.Nodes[0].KernelCmdLine
}

func (bundle *CrcBundleInfo) GetBundleBuildTime() (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(bundle.BuildInfo.BuildTime))
}

func (bundle *CrcBundleInfo) GetOpenshiftVersion() string {
	return bundle.ClusterInfo.OpenShiftVersion.String()
}

func (bundle *CrcBundleInfo) GetPodmanVersion() string {
	return bundle.Nodes[0].PodmanVersion
}

func (bundle *CrcBundleInfo) GetBundleNameWithoutExtension() string {
	return GetBundleNameWithoutExtension(bundle.GetBundleName())
}

func (bundle *CrcBundleInfo) getBundleType() Type {
	if bundle.Type == "snc" {
		return OpenShift
	}
	return Podman
}

func (bundle *CrcBundleInfo) IsOpenShift() bool {
	return bundle.getBundleType() == OpenShift
}

func (bundle *CrcBundleInfo) verify() error {
	files := []string{
		bundle.GetSSHKeyPath(),
		bundle.GetDiskImagePath(),
		bundle.GetKernelPath(),
		bundle.GetInitramfsPath(),
	}
	if bundle.IsOpenShift() {
		files = append(files, []string{
			bundle.GetOcPath(),
			bundle.GetKubeConfigPath()}...)
	}

	for _, file := range files {
		if file == "" {
			continue
		}
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%s not found in bundle", filepath.Base(file))
			}
			return err
		}
	}
	return bundle.checkDiskImageSize()
}

func (bundle *CrcBundleInfo) checkDiskImageSize() error {
	diskImagePath := bundle.GetDiskImagePath()
	expectedSize, err := strconv.ParseInt(bundle.Storage.DiskImages[0].Size, 10, 64)
	if err != nil {
		return err
	}
	stat, err := os.Stat(diskImagePath)
	if err != nil {
		return err
	}
	gotSize := stat.Size()
	if expectedSize != gotSize {
		return fmt.Errorf("unexpected disk image size: got %d instead of %d", gotSize, expectedSize)
	}
	return nil
}

func GetBundleNameWithoutExtension(bundleName string) string {
	return strings.TrimSuffix(bundleName, bundleExtension)
}

func GetCustomBundleName(bundleFilename string) string {
	re := regexp.MustCompile(`(?:_[0-9]+)*.crcbundle$`)
	baseName := re.ReplaceAllLiteralString(bundleFilename, "")
	return fmt.Sprintf("%s_%d%s", baseName, time.Now().Unix(), bundleExtension)
}

type bundlesDownloadInfo map[string]*download.RemoteFile

func getBundleDownloadInfo() (*download.RemoteFile, error) {
	bundles, ok := bundleLocations[runtime.GOARCH]
	if !ok {
		return nil, fmt.Errorf("Unsupported architecture: %s", runtime.GOARCH)
	}
	downloadInfo, ok := bundles[runtime.GOOS]
	if !ok {
		return nil, fmt.Errorf("Unknown GOOS: %s", runtime.GOOS)
	}

	return downloadInfo, nil
}

func Download() error {
	downloadInfo, err := getBundleDownloadInfo()
	if err != nil {
		return err
	}
	if _, err := downloadInfo.Download(constants.DefaultBundlePath, 0664); err != nil {
		return err
	}

	return nil
}
