package bundle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/extract"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

func getCachedBundlePath(cacheDir, bundleName string) string {
	path := strings.TrimSuffix(bundleName, ".crcbundle")
	return filepath.Join(cacheDir, path)
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

type Repository struct {
	CacheDir string
	OcBinDir string
}

func (repo *Repository) Get(bundleName string) (*CrcBundleInfo, error) {
	path := getCachedBundlePath(repo.CacheDir, bundleName)
	if _, err := os.Stat(path); err != nil {
		return nil, errors.Wrapf(err, "could not find cached bundle info in %s", path)
	}
	var bundleInfo CrcBundleInfo
	bundleInfo.cachedPath = path
	if err := bundleInfo.readBundleInfo(); err != nil {
		return nil, err
	}
	return &bundleInfo, nil
}

func (repo *Repository) Use(bundleName string) (*CrcBundleInfo, error) {
	bundleInfo, err := repo.Get(bundleName)
	if err != nil {
		return nil, err
	}
	if err := bundleInfo.createSymlinkOrCopyOpenShiftClient(repo.OcBinDir); err != nil {
		return nil, err
	}
	return bundleInfo, nil
}

func (bundle *CrcBundleInfo) createSymlinkOrCopyOpenShiftClient(ocBinDir string) error {
	ocInBundle := filepath.Join(bundle.cachedPath, constants.OcExecutableName)
	ocInBinDir := filepath.Join(ocBinDir, constants.OcExecutableName)
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
	_, err := extract.Uncompress(path, repo.CacheDir, true)
	return err
}

var defaultRepo = &Repository{
	CacheDir: constants.MachineCacheDir,
	OcBinDir: constants.CrcOcBinDir,
}

func GetCachedBundleInfo(bundleName string) (*CrcBundleInfo, error) {
	return defaultRepo.Use(bundleName)
}

func Extract(path string) (*CrcBundleInfo, error) {
	if err := defaultRepo.Extract(path); err != nil {
		return nil, err
	}
	return defaultRepo.Use(filepath.Base(path))
}
