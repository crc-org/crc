package bundle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/code-ready/crc/pkg/crc/constants"
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
	path := filepath.Join(repo.CacheDir, strings.TrimSuffix(bundleName, bundleExtension))
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
	if fmt.Sprintf(".%s", bundleInfo.ClusterInfo.AppsDomain) != constants.AppsDomain {
		return nil, fmt.Errorf("unexpected bundle, it must have %s apps domain", constants.AppsDomain)
	}
	if bundleInfo.GetAPIHostname() != fmt.Sprintf("api%s", constants.ClusterDomain) {
		return nil, fmt.Errorf("unexpected bundle, it must have %s base domain", constants.ClusterDomain)
	}
	if err := bundleInfo.createSymlinkOrCopyOpenShiftClient(repo.OcBinDir); err != nil {
		return nil, err
	}
	return bundleInfo, nil
}

func (bundle *CrcBundleInfo) createSymlinkOrCopyOpenShiftClient(ocBinDir string) error {
	ocInBundle := bundle.resolvePath(constants.OcExecutableName)
	ocInBinDir := filepath.Join(ocBinDir, constants.OcExecutableName)

	// this is needed when upgrading from crc versions anterior to commit 1.11.0~5
	if info, err := os.Stat(ocBinDir); err == nil && !info.IsDir() {
		if err := os.Remove(ocBinDir); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(ocBinDir, 0750); err != nil {
		return err
	}
	_ = os.Remove(ocInBinDir)
	if runtime.GOOS == "windows" {
		return crcos.CopyFileContents(ocInBundle, ocInBinDir, 0750)
	}
	return os.Symlink(ocInBundle, ocInBinDir)
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

	bundleDir := strings.TrimSuffix(bundleName, bundleExtension)
	return os.Rename(filepath.Join(tmpDir, bundleDir), filepath.Join(repo.CacheDir, bundleDir))
}

var defaultRepo = &Repository{
	CacheDir: constants.MachineCacheDir,
	OcBinDir: constants.CrcOcBinDir,
}

func Get(bundleName string) (*CrcBundleInfo, error) {
	return defaultRepo.Use(bundleName)
}

func Extract(path string) (*CrcBundleInfo, error) {
	if err := defaultRepo.Extract(path); err != nil {
		return nil, err
	}
	return defaultRepo.Use(filepath.Base(path))
}
