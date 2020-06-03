package cache

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/extract"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

type TrayCache struct {
	Cache
}

func (tc *TrayCache) CacheBinary() error {
	if tc.IsCached() {
		return nil
	}

	// Create tmp dir to download the requested tarball
	tmpDir, err := ioutil.TempDir("", "crc")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	assetTmpFile, err := tc.getBinary(tmpDir)
	if err != nil {
		return err
	}

	// Extract the tarball and put it the cache directory.
	extractedFiles, err := extract.Uncompress(assetTmpFile, tmpDir)
	if err != nil {
		return errors.Wrapf(err, "Cannot uncompress '%s'", assetTmpFile)
	}

	// Copy the requested asset into its final destination
	err = os.MkdirAll(tc.destDir, 0750)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "Cannot create the target directory.")
	}

	for _, extractedFilePath := range extractedFiles {
		finalBinaryPath := filepath.Join(tc.destDir, filepath.Base(extractedFilePath))
		err = crcos.CopyFileContents(extractedFilePath, finalBinaryPath, 0500)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewWindowsTrayCache(destDir string) *TrayCache {
	return &TrayCache{Cache{binaryName: constants.TrayBinaryName, archiveURL: constants.GetCRCWindowsTrayDownloadURL(), destDir: destDir}}
}
