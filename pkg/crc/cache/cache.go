package cache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

type Cache struct {
	executableName string
	archiveURL     string
	destDir        string
	version        string
	getVersion     func(string) (string, error)
}

type VersionMismatchError struct {
	ExecutableName  string
	ExpectedVersion string
	CurrentVersion  string
}

func (e *VersionMismatchError) Error() string {
	return fmt.Sprintf("%s version mismatch: %s expected but %s found in the cache", e.ExecutableName, e.ExpectedVersion, e.CurrentVersion)
}

func New(executableName string, archiveURL string, destDir string, version string, getVersion func(string) (string, error)) *Cache {
	return &Cache{executableName: executableName, archiveURL: archiveURL, destDir: destDir, version: version, getVersion: getVersion}
}

func (c *Cache) GetExecutablePath() string {
	return filepath.Join(c.destDir, c.executableName)
}

func (c *Cache) GetExecutableName() string {
	return c.executableName
}

/* getVersionGeneric runs the cached executable with 'args', and assumes the version string
 * was output on stdout's first line with this format:
 * something: <version>
 *
 * It returns <version> as a string
 */
func getVersionGeneric(executablePath string, args ...string) (string, error) { //nolint:deadcode,unused
	stdOut, _, err := crcos.RunWithDefaultLocale(executablePath, args...)
	if err != nil {
		return "", err
	}
	parsedOutput := strings.Split(stdOut, ":")
	if len(parsedOutput) < 2 {
		return "", fmt.Errorf("Unable to parse the version information of %s", executablePath)
	}
	return strings.TrimSpace(parsedOutput[1]), nil
}

func NewPodmanCache() *Cache {
	return New(constants.PodmanExecutableName, constants.GetPodmanURL(), constants.CrcBinDir, "", nil)
}

func NewGoodhostsCache() *Cache {
	return New(constants.GoodhostsExecutableName, constants.GetGoodhostsURL(), constants.CrcBinDir, "", nil)
}

func (c *Cache) IsCached() bool {
	if _, err := os.Stat(c.GetExecutablePath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *Cache) EnsureIsCached() error {
	if !c.IsCached() || c.CheckVersion() != nil {
		err := c.CacheExecutable()
		if err != nil {
			return err
		}
	}
	return nil
}

// CacheExecutable downloads and caches the requested executable into the CRC directory
func (c *Cache) CacheExecutable() error {
	if c.IsCached() && c.CheckVersion() == nil {
		return nil
	}

	// Create tmp dir to download the requested tarball
	tmpDir, err := ioutil.TempDir("", "crc")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	assetTmpFile, err := c.getExecutable(tmpDir)
	if err != nil {
		return err
	}

	var extractedFiles []string
	// Check the file is tarball or not
	if IsTarball(assetTmpFile) {
		// Extract the tarball and put it the cache directory.
		extractedFiles, err = extract.UncompressWithFilter(assetTmpFile, tmpDir, false,
			func(filename string) bool { return filepath.Base(filename) == c.executableName })
		if err != nil {
			return errors.Wrapf(err, "Cannot uncompress '%s'", assetTmpFile)
		}
	} else {
		extractedFiles = append(extractedFiles, assetTmpFile)
		if filepath.Base(assetTmpFile) != c.executableName {
			logging.Warnf("Executable name is %s but extracted file name is %s", c.executableName, filepath.Base(assetTmpFile))
		}
	}

	// Copy the requested asset into its final destination
	err = os.MkdirAll(c.destDir, 0750)
	if err != nil {
		return errors.Wrap(err, "Cannot create the target directory.")
	}

	for _, extractedFilePath := range extractedFiles {
		finalExecutablePath := filepath.Join(c.destDir, filepath.Base(extractedFilePath))
		// If the file exists then remove it (ignore error) first before copy because with `0500` permission
		// it is not possible to overwrite the file.
		os.Remove(finalExecutablePath)
		err = crcos.CopyFileContents(extractedFilePath, finalExecutablePath, 0500)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) getExecutable(destDir string) (string, error) {
	logging.Debugf("Trying to extract %s from crc executable", c.executableName)
	archiveName := filepath.Base(c.archiveURL)
	destPath := filepath.Join(destDir, archiveName)
	err := embed.Extract(archiveName, destPath)
	if err != nil {
		return download.Download(c.archiveURL, destDir, 0600)
	}

	return destPath, err
}

func (c *Cache) CheckVersion() error {
	// Check if version string is non-empty
	if c.version == "" {
		return nil
	}
	currentVersion, err := c.getVersion(c.GetExecutablePath())
	if err != nil {
		return err
	}
	if currentVersion != c.version {
		err := &VersionMismatchError{
			ExecutableName:  c.executableName,
			CurrentVersion:  currentVersion,
			ExpectedVersion: c.version,
		}
		logging.Debugf("%s", err.Error())
		return err
	}
	logging.Debugf("Found %s version %s", c.executableName, c.version)
	return nil
}

func IsTarball(filename string) bool {
	tarballExtensions := []string{".tar", ".tar.gz", ".tar.xz", ".zip", ".tar.bz2", ".crcbundle"}
	for _, extension := range tarballExtensions {
		if strings.HasSuffix(strings.ToLower(filename), extension) {
			return true
		}
	}
	return false
}
