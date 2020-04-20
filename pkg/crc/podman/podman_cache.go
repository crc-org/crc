package podman

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

// Podman is a struct with methods designed for dealing with the podman binary
type PodmanCached struct{}

func (podman *PodmanCached) EnsureIsCached() error {
	if !podman.IsCached() {
		err := podman.cachePodman()
		if err != nil {
			return err
		}

	}
	return nil
}

func (podman *PodmanCached) IsCached() bool {
	if _, err := os.Stat(filepath.Join(constants.CrcBinDir, constants.PodmanBinaryName)); os.IsNotExist(err) {
		return false
	}
	return true
}

func matchPodmanBinaryName(filename string) bool {
	return filepath.Base(filename) == constants.PodmanBinaryName
}

func (podman *PodmanCached) getPodman(destDir string) (string, error) {
	logging.Debug("Trying to extract podman from crc binary")
	archiveName := filepath.Base(constants.GetPodmanUrl())
	destPath := filepath.Join(destDir, archiveName)
	err := embed.Extract(archiveName, destPath)
	if err != nil {
		logging.Debug("Downloading podman")
		return download.Download(constants.GetPodmanUrl(), destDir, 0600)
	}

	return destPath, err
}

// cachePodman downloads and caches the podman binary into the CRC directory
func (podman *PodmanCached) cachePodman() error {
	if podman.IsCached() {
		return nil
	}

	// Create tmp dir to download the podman tarball
	tmpDir, err := ioutil.TempDir("", "crc")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	assetTmpFile, err := podman.getPodman(tmpDir)
	if err != nil {
		return err
	}

	// Extract the tarball and put it the cache directory.
	extractedFiles, err := extract.UncompressWithFilter(assetTmpFile, tmpDir, matchPodmanBinaryName)
	if err != nil {
		return errors.Wrapf(err, "Cannot uncompress '%s'", assetTmpFile)
	}

	// Copy the requested asset into its final destination
	outputPath := constants.CrcBinDir
	err = os.MkdirAll(outputPath, 0750)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "Cannot create the target directory.")
	}

	for _, extractedFilePath := range extractedFiles {
		finalBinaryPath := filepath.Join(outputPath, filepath.Base(extractedFilePath))
		err = crcos.CopyFileContents(extractedFilePath, finalBinaryPath, 0500)
		if err != nil {
			return err
		}
	}

	return nil
}
