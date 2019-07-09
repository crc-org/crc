package oc

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/extract"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

const (
	OC_CACHE_DIR = "oc"
	TAR          = "tar.gz"
	ZIP          = "zip"
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

// cacheOc downloads and caches the oc binary into the minishift directory
func (oc *OcCached) cacheOc() error {
	if !oc.IsCached() {
		logging.Debug("Downloading oc")
		// Create tmp dir to download the oc tarball
		tmpDir, err := ioutil.TempDir("", "crc")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)
		assetTmpFile, err := download.Download(constants.DefaultOcURL, tmpDir, 0600)
		if err != nil {
			return err
		}

		// Extract the tarball and put it the cache directory.
		binaryPath := ""
		switch {
		case strings.HasSuffix(assetTmpFile, TAR):
			// unzip
			tarFile := assetTmpFile[:len(assetTmpFile)-3]
			err = extract.Ungzip(assetTmpFile, tarFile)
			if err != nil {
				return errors.Wrapf(err, "Cannot ungzip '%s'", assetTmpFile)
			}

			// untar
			err = extract.Untar(tarFile, tmpDir)
			if err != nil {
				return errors.Wrapf(err, "Cannot untar '%s'", tarFile)
			}

			content, err := listDirExcluding(tmpDir, ".*.tar.*")
			if err != nil {
				return errors.Wrapf(err, "Cannot list content of '%s'", tmpDir)
			}
			if len(content) > 1 {
				return errors.New(fmt.Sprintf("Unexpected number of files in tmp directory: %s", content))
			}

			binaryPath = tmpDir
		case strings.HasSuffix(assetTmpFile, ZIP):
			contentDir := assetTmpFile[:len(assetTmpFile)-4]
			err = extract.Unzip(assetTmpFile, contentDir)
			if err != nil {
				return errors.Wrapf(err, "Cannot unzip '%s'", assetTmpFile)
			}
			binaryPath = contentDir
		}

		binaryName := constants.OcBinaryName
		binaryPath = filepath.Join(binaryPath, binaryName)

		// Copy the requested asset into its final destination
		outputPath := constants.CrcBinDir
		err = os.MkdirAll(outputPath, 0755)
		if err != nil && !os.IsExist(err) {
			return errors.Wrap(err, "Cannot create the target directory.")
		}

		finalBinaryPath := filepath.Join(outputPath, binaryName)
		crcos.CopyFileContents(binaryPath, finalBinaryPath, 0755)
		if err != nil {
			return err
		}

		err = os.Chmod(finalBinaryPath, 0777)
		if err != nil {
			return errors.Wrapf(err, "Cannot make '%s' executable", finalBinaryPath)
		}

		return nil
	}
	return nil
}

func listDirExcluding(dir string, excludeRegexp string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, f := range files {
		matched, err := regexp.MatchString(excludeRegexp, f.Name())
		if err != nil {
			return nil, err
		}

		if !matched {
			result = append(result, f.Name())
		}

	}

	return result, nil
}
