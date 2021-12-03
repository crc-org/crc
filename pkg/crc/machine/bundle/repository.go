package bundle

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/extract"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

const (
	supportedVersion = "^1.0"
	bundleExtension  = ".crcbundle"
	metadataFilename = "crc-bundle-info.json"
)

type Repository struct {
	CacheDir string
	OcBinDir string
}

func (repo *Repository) Get(bundleName string) (*CrcBundleInfo, error) {
	path := filepath.Join(repo.CacheDir, GetBundleNameWithoutExtension(bundleName))
	if _, err := os.Stat(path); err != nil {
		return nil, errors.Wrapf(err, "could not find cached bundle info in %s", path)
	}
	jsonFilepath := filepath.Join(path, metadataFilename)
	content, err := ioutil.ReadFile(jsonFilepath)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading %s file", jsonFilepath)
	}
	var bundleInfo CrcBundleInfo
	if err := json.Unmarshal(content, &bundleInfo); err != nil {
		return nil, errors.Wrap(err, "error Unmarshal the data")
	}
	if err := checkVersion(bundleInfo); err != nil {
		return nil, err
	}
	bundleInfo.cachedPath = path

	if err := bundleInfo.verify(); err != nil {
		return nil, err
	}

	if bundleInfo.IsOpenShift() {
		if fmt.Sprintf(".%s", bundleInfo.ClusterInfo.AppsDomain) != constants.AppsDomain {
			return nil, fmt.Errorf("unexpected bundle, it must have %s apps domain", constants.AppsDomain)
		}
		if bundleInfo.GetAPIHostname() != fmt.Sprintf("api%s", constants.ClusterDomain) {
			return nil, fmt.Errorf("unexpected bundle, it must have %s base domain", constants.ClusterDomain)
		}
	}
	return &bundleInfo, nil
}

func checkVersion(bundleInfo CrcBundleInfo) error {
	version, err := semver.NewVersion(bundleInfo.Version)
	if err != nil {
		return errors.Wrap(err, "cannot parse bundle version")
	}
	constraint, err := semver.NewConstraint(supportedVersion)
	if err != nil {
		return errors.Wrap(err, "cannot parse version constraint")
	}
	if !constraint.Check(version) {
		return fmt.Errorf("cannot use bundle with version %s, bundle version must satisfy %s constraint", bundleInfo.Version, supportedVersion)
	}
	return nil
}

func (repo *Repository) Use(bundleName string) (*CrcBundleInfo, error) {
	bundleInfo, err := repo.Get(bundleName)
	if err != nil {
		return nil, err
	}
	if bundleInfo.IsOpenShift() {
		if err := bundleInfo.createSymlinkOrCopyOpenShiftClient(repo.OcBinDir); err != nil {
			return nil, err
		}
	}
	if err := bundleInfo.createSymlinkOrCopyPodmanClient(repo.OcBinDir); err != nil {
		return nil, err
	}
	return bundleInfo, nil
}

func (bundle *CrcBundleInfo) copyExecutableFromBundle(destDir string, fileType FileType, executableName string) error {
	srcPath := bundle.getHelperPath(fileType)
	if srcPath == "" {
		// this can happen if the bundle metadata does not list an executable for 'fileType'
		return nil
	}
	destPath := filepath.Join(destDir, executableName)

	if err := os.MkdirAll(destDir, 0750); err != nil {
		return err
	}
	_ = os.Remove(destPath)
	if runtime.GOOS == "windows" {
		return crcos.CopyFileContents(srcPath, destPath, 0750)
	}
	// on unix-like OSes, a symlink is good enough, no need for a copy
	return os.Symlink(srcPath, destPath)
}

func (bundle *CrcBundleInfo) createSymlinkOrCopyOpenShiftClient(ocBinDir string) error {
	return bundle.copyExecutableFromBundle(ocBinDir, OcExecutable, constants.OcExecutableName)
}

func (bundle *CrcBundleInfo) createSymlinkOrCopyPodmanClient(binDir string) error {
	return bundle.copyExecutableFromBundle(binDir, PodmanExecutable, constants.PodmanRemoteExecutableName)
}

func (repo *Repository) Extract(path string) error {
	bundleName := filepath.Base(path)

	tmpDir := filepath.Join(repo.CacheDir, "tmp-extract")
	_ = os.RemoveAll(tmpDir) // clean up before using it
	defer func() {
		_ = os.RemoveAll(tmpDir) // clean up after using it
	}()

	if _, err := extract.Uncompress(path, tmpDir, true); err != nil {
		return err
	}

	bundleBaseDir := GetBundleNameWithoutExtension(bundleName)
	bundleDir := filepath.Join(repo.CacheDir, bundleBaseDir)
	_ = os.RemoveAll(bundleDir)
	err := crcerrors.Retry(context.Background(), time.Minute, func() error {
		if err := os.Rename(filepath.Join(tmpDir, bundleBaseDir), bundleDir); err != nil {
			return &crcerrors.RetriableError{Err: err}
		}
		return nil
	}, 5*time.Second)
	if err != nil {
		return err
	}

	return os.Chmod(bundleDir, 0755)
}

func (repo *Repository) List() ([]CrcBundleInfo, error) {
	files, err := ioutil.ReadDir(repo.CacheDir)
	if err != nil {
		return nil, err
	}
	var ret []CrcBundleInfo
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		bundle, err := repo.Get(file.Name())
		if err != nil {
			logging.Errorf("cannot load bundle %s: %v", file.Name(), err)
			continue
		}
		ret = append(ret, *bundle)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].ClusterInfo.OpenShiftVersion.GreaterThan(ret[j].ClusterInfo.OpenShiftVersion)
	})
	return ret, nil
}

func (repo *Repository) CalculateBundleSha256Sum(bundlePath string) (string, error) {
	return sha256sum(bundlePath)
}

var defaultRepo = &Repository{
	CacheDir: constants.MachineCacheDir,
	OcBinDir: constants.CrcOcBinDir,
}

func Get(bundleName string) (*CrcBundleInfo, error) {
	return defaultRepo.Get(bundleName)
}

func CalculateBundleSha256Sum(bundlePath string) (string, error) {
	return defaultRepo.CalculateBundleSha256Sum(bundlePath)
}

func Use(bundleName string) (*CrcBundleInfo, error) {
	return defaultRepo.Use(bundleName)
}

func Extract(path string) (*CrcBundleInfo, error) {
	if err := defaultRepo.Extract(path); err != nil {
		return nil, err
	}
	return defaultRepo.Get(filepath.Base(path))
}

func List() ([]CrcBundleInfo, error) {
	return defaultRepo.List()
}
