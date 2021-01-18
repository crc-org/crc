// +build !linux

package preflight

import (
	"io/ioutil"
	goos "os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/logging"
	dl "github.com/code-ready/crc/pkg/download"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/code-ready/crc/pkg/extract"
	"github.com/pkg/errors"
)

// Extracts or downloads the tray and puts it in the 'destDir' directory
func downloadOrExtractTrayApp(url, destDir string) error {
	logging.Debug("Downloading/extracting tray executable")

	tmpArchivePath, err := ioutil.TempDir("", "crc")
	if err != nil {
		logging.Error("Failed creating temporary directory for extracting tray")
		return err
	}
	defer func() {
		_ = goos.RemoveAll(tmpArchivePath)
	}()

	logging.Debug("Trying to extract tray from crc executable")
	trayFileName := filepath.Base(url)
	trayDestFileName := filepath.Join(tmpArchivePath, trayFileName)
	err = embed.Extract(trayFileName, trayDestFileName)
	if err != nil {
		logging.Debug("Could not extract tray from crc executable", err)
		logging.Debug("Downloading crc tray")
		_, err = dl.Download(url, tmpArchivePath, 0600)
		if err != nil {
			return err
		}
	}
	err = goos.MkdirAll(destDir, 0750)
	if err != nil {
		return errors.Wrap(err, "Cannot create the target directory.")
	}
	_, err = extract.Uncompress(trayDestFileName, destDir, false)
	if err != nil {
		return errors.Wrapf(err, "Cannot uncompress '%s'", trayDestFileName)
	}
	return nil
}
