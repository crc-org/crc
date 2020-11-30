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
)

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
	if err := bundleInfo.createSymlinkOrCopyOpenShiftClient(); err != nil {
		return nil, err
	}
	return &bundleInfo, nil
}

func (bundle *CrcBundleInfo) createSymlinkOrCopyOpenShiftClient() error {
	ocInBundle := filepath.Join(bundle.cachedPath, constants.OcExecutableName)
	ocInBinDir := filepath.Join(constants.CrcOcBinDir, constants.OcExecutableName)
	if err := os.MkdirAll(constants.CrcOcBinDir, 0750); err != nil {
		return err
	}
	_ = os.Remove(ocInBinDir)
	if runtime.GOOS == "windows" {
		return crcos.CopyFileContents(ocInBundle, ocInBinDir, 0750)
	}
	return os.Symlink(ocInBundle, ocInBinDir)
}

func Extract(sourcepath string) (*CrcBundleInfo, error) {
	_, err := extract.Uncompress(sourcepath, constants.MachineCacheDir, true)
	if err != nil {
		return nil, err
	}

	return GetCachedBundleInfo(filepath.Base(sourcepath))
}
