package oc

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

// Oc is a struct with methods designed for dealing with the oc binary
type OcCached struct{}

func (oc *OcCached) EnsureIsCached() error {
	if !oc.IsCached() {
		err := oc.cacheOc()
		if err != nil {
			return err
		}

	}
	return nil
}

func (oc *OcCached) IsCached() bool {
	if _, err := os.Stat(filepath.Join(constants.CrcBinDir, constants.OcBinaryName)); os.IsNotExist(err) {
		return false
	}
	return true
}

func matchOcBinaryName(filename string) bool {
	return filepath.Base(filename) == constants.OcBinaryName
}

func (oc *OcCached) getOc(destDir string) (string, error) {
	logging.Debug("Trying to extract oc from crc binary")
	archiveName := filepath.Base(constants.GetOcUrl())
	destPath := filepath.Join(destDir, archiveName)
	err := embed.Extract(archiveName, destPath)
	if err != nil {
		logging.Debug("Downloading oc")
		return download.Download(constants.GetOcUrl(), destDir, 0600)
	}

	return destPath, err
}

// cacheOc downloads and caches the oc binary into the minishift directory
func (oc *OcCached) cacheOc() error {
	if !oc.IsCached() {
		// Create tmp dir to download the oc tarball
		tmpDir, err := ioutil.TempDir("", "crc")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)
		assetTmpFile, err := oc.getOc(tmpDir)
		if err != nil {
			return err
		}

		// Extract the tarball and put it the cache directory.
		_, err = extract.UncompressWithFilter(assetTmpFile, tmpDir, matchOcBinaryName)
		if err != nil {
			return errors.Wrapf(err, "Cannot uncompress '%s'", assetTmpFile)
		}

		binaryName := constants.OcBinaryName
		binaryPath := filepath.Join(tmpDir, binaryName)

		// Copy the requested asset into its final destination
		outputPath := constants.CrcBinDir
		err = os.MkdirAll(outputPath, 0750)
		if err != nil && !os.IsExist(err) {
			return errors.Wrap(err, "Cannot create the target directory.")
		}

		finalBinaryPath := filepath.Join(outputPath, binaryName)
		err = crcos.CopyFileContents(binaryPath, finalBinaryPath, 0500)
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}
