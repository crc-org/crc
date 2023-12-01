package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/version"
	"github.com/crc-org/crc/v2/pkg/download"
	"github.com/crc-org/crc/v2/pkg/embed"
	"github.com/crc-org/crc/v2/pkg/extract"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/pkg/errors"
)

type Cache struct {
	executablePath     string
	archiveURL         string
	version            string
	ignoreNameMismatch bool
	getVersion         func(string) (string, error)
}

type VersionMismatchError struct {
	ExecutableName  string
	ExpectedVersion string
	CurrentVersion  string
}

func (e *VersionMismatchError) Error() string {
	return fmt.Sprintf("%s version mismatch: %s expected but %s found in the cache", e.ExecutableName, e.ExpectedVersion, e.CurrentVersion)
}

func newCache(executablePath string, archiveURL string, version string, getVersion func(string) (string, error)) *Cache {
	return &Cache{executablePath: executablePath, archiveURL: archiveURL, version: version, getVersion: getVersion}
}

func (c *Cache) GetExecutablePath() string {
	return c.executablePath
}

func (c *Cache) GetExecutableName() string {
	return filepath.Base(c.executablePath)
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
		logging.Debugf("failed to run executable %s: %v", executablePath, err)
		return "", err
	}
	parsedOutput := strings.Split(stdOut, ":")
	if len(parsedOutput) < 2 {
		logging.Debugf("failed to parse version information for %s: %s", executablePath, stdOut)
		return "", fmt.Errorf("Unable to parse the version information of %s", executablePath)
	}
	return strings.TrimSpace(parsedOutput[1]), nil
}

func NewAdminHelperCache() *Cache {
	url := constants.GetAdminHelperURL()
	version := version.GetAdminHelperVersion()
	return newCache(constants.AdminHelperPath(),
		url,
		version,
		func(executable string) (string, error) {
			out, _, err := crcos.RunWithDefaultLocale(executable, "--version")
			if err != nil {
				return "", err
			}
			split := strings.Split(out, " ")
			return strings.TrimSpace(split[len(split)-1]), nil
		},
	)
}

func (c *Cache) IsCached() bool {
	if _, err := os.Stat(c.GetExecutablePath()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *Cache) EnsureIsCached() error {
	if !c.IsCached() || c.CheckVersion() != nil {
		if version.IsInstaller() {
			return fmt.Errorf("%s could not be found - check your installation", c.GetExecutablePath())
		}
		return c.cacheExecutable()
	}
	return nil
}

// CacheExecutable downloads and caches the requested executable into the CRC directory
func (c *Cache) cacheExecutable() error {
	if c.IsCached() && c.CheckVersion() == nil {
		return nil
	}

	// Create tmp dir to download the requested tarball
	tmpDir, err := os.MkdirTemp("", "crc")
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
	if isTarball(assetTmpFile) {
		// Extract the tarball and put it the cache directory.
		extractedFiles, err = extract.UncompressWithFilter(assetTmpFile, tmpDir,
			func(filename string) bool { return filepath.Base(filename) == c.GetExecutableName() })
		if err != nil {
			return errors.Wrapf(err, "Cannot uncompress '%s'", assetTmpFile)
		}
	} else {
		extractedFiles = append(extractedFiles, assetTmpFile)
		if filepath.Base(assetTmpFile) != c.GetExecutableName() && !c.ignoreNameMismatch {
			logging.Warnf("Executable name is %s but extracted file name is %s", c.GetExecutableName(), filepath.Base(assetTmpFile))
		}
	}

	// Copy the requested asset into its final destination
	for _, extractedFilePath := range extractedFiles {
		finalExecutablePath := filepath.Join(constants.CrcBinDir, c.GetExecutableName())
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
	logging.Debugf("Trying to extract %s from crc executable", c.GetExecutableName())
	archiveName := filepath.Base(c.archiveURL)
	destPath := filepath.Join(destDir, archiveName)
	err := embed.Extract(archiveName, destPath)
	if err != nil {
		return download.Download(c.archiveURL, destDir, 0600, nil)
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
			ExecutableName:  c.GetExecutableName(),
			CurrentVersion:  currentVersion,
			ExpectedVersion: c.version,
		}
		logging.Debugf("%s", err.Error())
		return err
	}
	logging.Debugf("Found %s version %s", c.GetExecutableName(), c.version)
	return nil
}

func isTarball(filename string) bool {
	tarballExtensions := []string{".tar", ".tar.gz", ".tar.xz", ".zip", ".tar.bz2", ".crcbundle"}
	for _, extension := range tarballExtensions {
		if strings.HasSuffix(strings.ToLower(filename), extension) {
			return true
		}
	}
	return false
}
