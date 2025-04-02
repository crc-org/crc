package bundle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/gpg"
	"github.com/crc-org/crc/v2/pkg/crc/image"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/download"
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

type FilenameInfo struct {
	Preset             crcPreset.Preset
	Driver             string
	Version            string
	Arch               string
	CustomBundleSuffix string
}

type Node struct {
	Kind          []string `json:"kind"`
	Hostname      string   `json:"hostname"`
	DiskImage     string   `json:"diskImage"`
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

func (bundle *CrcBundleInfo) GetFQDN(shortName string) string {
	return fmt.Sprintf("%s.%s.%s", shortName, bundle.ClusterInfo.ClusterName, bundle.ClusterInfo.BaseDomain)
}

func (bundle *CrcBundleInfo) GetAPIHostname() string {
	return bundle.GetFQDN("api")
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

func (bundle *CrcBundleInfo) getHelperPath(fileType FileType) string {
	for _, file := range bundle.Storage.Files {
		if file.Type == fileType {
			return bundle.resolvePath(file.Name)
		}
	}

	return ""
}

func (bundle *CrcBundleInfo) GetOcPath() string {
	return bundle.getHelperPath(OcExecutable)
}

func (bundle *CrcBundleInfo) GetPodmanPath() string {
	return bundle.getHelperPath(PodmanExecutable)
}

func (bundle *CrcBundleInfo) GetSSHKeyPath() string {
	return bundle.resolvePath(bundle.ClusterInfo.SSHPrivateKeyFile)
}

func (bundle *CrcBundleInfo) GetBundleBuildTime() (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(bundle.BuildInfo.BuildTime))
}

func (bundle *CrcBundleInfo) GetVersion() string {
	return bundle.ClusterInfo.OpenShiftVersion.String()
}

func (bundle *CrcBundleInfo) GetPodmanVersion() string {
	return bundle.Nodes[0].PodmanVersion
}

func (bundle *CrcBundleInfo) GetBundleNameWithoutExtension() string {
	return GetBundleNameWithoutExtension(bundle.GetBundleName())
}

func (bundle *CrcBundleInfo) GetBundleType() crcPreset.Preset {
	bundleType := strings.TrimSuffix(bundle.Type, "_custom")
	if bundleType == "snc" {
		bundleType = "openshift"
	}
	return crcPreset.ParsePreset(bundleType)
}

func (bundle *CrcBundleInfo) IsOpenShift() bool {
	preset := bundle.GetBundleType()
	return preset == crcPreset.OpenShift || preset == crcPreset.OKD
}

func (bundle *CrcBundleInfo) IsMicroshift() bool {
	return bundle.GetBundleType() == crcPreset.Microshift
}

func (bundle *CrcBundleInfo) verify() error {
	files := []string{
		bundle.GetSSHKeyPath(),
		bundle.GetDiskImagePath(),
		bundle.GetOcPath(),
		bundle.GetKubeConfigPath()}

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

func GetBundleNameWithExtension(bundleName string) string {
	if strings.HasSuffix(bundleName, bundleExtension) {
		return bundleName
	}
	return fmt.Sprintf("%s%s", bundleName, bundleExtension)
}

func GetCustomBundleName(bundleFilename string) string {
	re := regexp.MustCompile(`(?:_[0-9]+)*.crcbundle$`)
	baseName := re.ReplaceAllLiteralString(bundleFilename, "")
	return fmt.Sprintf("%s_%d%s", baseName, time.Now().Unix(), bundleExtension)
}

func GetBundleNameFromURI(bundleURI string) (string, error) {
	switch {
	case strings.HasPrefix(bundleURI, "docker://"):
		imageAndTag := strings.Split(path.Base(bundleURI), ":")
		if len(imageAndTag) < 2 {
			return "", fmt.Errorf("no tag found in bundle URI")
		}
		return constants.BundleForPreset(image.GetPresetName(imageAndTag[0]), imageAndTag[1]), nil
	case strings.HasPrefix(bundleURI, "http://"), strings.HasPrefix(bundleURI, "https://"):
		return path.Base(bundleURI), nil
	default:
		// local path
		return filepath.Base(bundleURI), nil
	}
}

// GetBundleInfoFromName Parses the bundle filename and returns a FilenameInfo struct
func GetBundleInfoFromName(bundleName string) (*FilenameInfo, error) {
	var filenameInfo FilenameInfo

	/*
		crc_preset_driver_version_arch_customSuffix.crcbundle

		crc                      : Matches the fixed crc part
		(?:(?:_)([[:alpha:]]+))? : Matches the preset part (optional)
		([[:alpha:]]+)           : Matches the next mandatory alphabetic part (e.g., libvirt)
		(%s = semverRegex)       : Matches the version in SemVer format (e.g., 4.16.7 or 4.16.7-ec.2)
		([[:alnum:]]+)           : Matches the architecture or platform part (e.g. amd64)
		(?:(?:_)([0-9]+))?       : Optionally matches a trailing number after an underscore (e.g. 2345).
		\.crcbundle              : Matches the file extension .crcbundle
	*/
	semverRegex := "(?:0|[1-9]\\d*)\\.(?:0|[1-9]\\d*)\\.(?:0|[1-9]\\d*)(?:-(?:(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?:[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?"
	bundleRegex := `crc(?:(?:_)([[:alpha:]]+))?_([[:alpha:]]+)_(%s)_([[:alnum:]]+)(?:(?:_)([0-9]+))?\.crcbundle`
	compiledRegex := regexp.MustCompile(fmt.Sprintf(bundleRegex, semverRegex))
	filenameParts := compiledRegex.FindStringSubmatch(bundleName)

	if filenameParts == nil {
		return &filenameInfo, fmt.Errorf("bundle filename is in unrecognized format")
	}

	if filenameParts[1] == "" {
		filenameInfo.Preset = crcPreset.OpenShift
	} else {
		parsedPreset, err := crcPreset.ParsePresetE(filenameParts[1])
		if err != nil {
			return &filenameInfo, err
		}
		filenameInfo.Preset = parsedPreset
	}
	filenameInfo.Driver = filenameParts[2]
	filenameInfo.Version = filenameParts[3]
	filenameInfo.Arch = filenameParts[4]
	filenameInfo.CustomBundleSuffix = filenameParts[5]

	return &filenameInfo, nil
}

func getBundleDownloadInfo(preset crcPreset.Preset) (*download.RemoteFile, error) {
	sha256sum, err := getDefaultBundleVerifiedHash(preset)
	if err != nil {
		return nil, fmt.Errorf("unable to get verified hash for default bundle: %w", err)
	}
	downloadInfo := download.NewRemoteFile(constants.GetDefaultBundleDownloadURL(preset), sha256sum)
	return downloadInfo, nil
}

// getDefaultBundleVerifiedHash downloads the sha256sum.txt.sig file from mirror.openshift.com
// then verifies it is signed by redhat release key, if signature is valid it returns the hash
// for the default bundle of preset from the file
func getDefaultBundleVerifiedHash(preset crcPreset.Preset) (string, error) {
	return getVerifiedHash(constants.GetDefaultBundleSignedHashURL(preset), constants.GetDefaultBundle(preset))
}

func getVerifiedHash(url string, file string) (string, error) {
	res, err := download.InMemory(url)
	if err != nil {
		return "", err
	}
	defer res.Close()
	signedHashes, err := io.ReadAll(res)
	if err != nil {
		return "", err
	}

	verifiedHashes, err := gpg.GetVerifiedClearsignedMsgV3(constants.RedHatReleaseKey, string(signedHashes))
	if err != nil {
		return "", fmt.Errorf("invalid signature: %w", err)
	}

	logging.Debugf("Verified bundle hashes:\n%s", verifiedHashes)

	lines := strings.Split(verifiedHashes, "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, file) {
			sha256sum := strings.TrimSuffix(line, "  "+file)
			return sha256sum, nil
		}
	}
	return "", fmt.Errorf("%s hash is missing or shasums are malformed", file)
}

func downloadDefault(ctx context.Context, preset crcPreset.Preset) (string, error) {
	downloadInfo, err := getBundleDownloadInfo(preset)
	if err != nil {
		return "", err
	}
	return downloadInfo.Download(ctx, constants.GetDefaultBundlePath(preset), 0664)
}

func Download(ctx context.Context, preset crcPreset.Preset, bundleURI string, enableBundleQuayFallback bool) (string, error) {
	// If we are asked to download
	// ~/.crc/cache/crc_podman_libvirt_4.1.1.crcbundle, this means we want
	// are downloading the default bundle for this release. This uses a
	// different codepath from user-specified URIs as for the default
	// bundles, their sha256sums are known and can be checked.
	if bundleURI == constants.GetDefaultBundlePath(preset) {
		switch preset {
		case crcPreset.OpenShift, crcPreset.Microshift:
			downloadedBundlePath, err := downloadDefault(ctx, preset)
			if err != nil && enableBundleQuayFallback {
				logging.Info("Unable to download bundle from mirror, falling back to quay")
				return image.PullBundle(ctx, constants.GetDefaultBundleImageRegistry(preset))
			}
			return downloadedBundlePath, err
		case crcPreset.OKD:
			fallthrough
		default:
			return image.PullBundle(ctx, constants.GetDefaultBundleImageRegistry(preset))
		}
	}
	switch {
	case strings.HasPrefix(bundleURI, "http://"), strings.HasPrefix(bundleURI, "https://"):
		return download.Download(ctx, bundleURI, constants.MachineCacheDir, 0644, nil)
	case strings.HasPrefix(bundleURI, "docker://"):
		return image.PullBundle(ctx, bundleURI)
	}
	// the `bundleURI` parameter turned out to be a local path
	return bundleURI, nil
}

type Version struct {
	CrcVersion       *semver.Version `json:"crcVersion"`
	GitSha           string          `json:"gitSha"`
	OpenshiftVersion string          `json:"openshiftVersion"`
}

type ReleaseInfo struct {
	Version Version           `json:"version"`
	Links   map[string]string `json:"links"`
}

func FetchLatestReleaseInfo() (*ReleaseInfo, error) {
	const releaseInfoLink = "https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/crc/latest/release-info.json"
	response, err := download.InMemory(releaseInfoLink)
	if err != nil {
		return nil, err
	}
	defer response.Close()

	releaseMetaData, err := io.ReadAll(response)
	if err != nil {
		return nil, err
	}

	var releaseInfo ReleaseInfo
	if err := json.Unmarshal(releaseMetaData, &releaseInfo); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON metadata: %v", err)
	}

	return &releaseInfo, nil
}
