package preflight

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/adminhelper"
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/embed"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

var bundleCheck = Check{
	configKeySuffix:  "check-bundle-extracted",
	checkDescription: "Checking if CRC bundle is extracted in '$HOME/.crc'",
	check:            checkBundleExtracted,
	fixDescription:   "Extracting bundle from the CRC executable",
	fix:              fixBundleExtracted,
	flags:            SetupOnly,

	labels: None,
}

var genericCleanupChecks = []Check{
	{
		cleanupDescription: "Removing CRC Machine Instance directory",
		cleanup:            removeCRCMachinesDir,
		flags:              CleanUpOnly,

		labels: None,
	},
	{
		cleanupDescription: "Removing older logs",
		cleanup:            removeOldLogs,
		flags:              CleanUpOnly,

		labels: None,
	},
	{
		cleanupDescription: "Removing pull secret from the keyring",
		cleanup:            cluster.ForgetPullSecret,
		flags:              CleanUpOnly,

		labels: None,
	},
	{
		cleanupDescription: "Removing hosts file records added by CRC",
		cleanup:            removeHostsFileEntry,
		flags:              CleanUpOnly,

		labels: None,
	},
}

func checkBundleExtracted() error {
	if !constants.IsRelease() {
		logging.Debugf("Development build, skipping check")
		return nil
	}
	logging.Infof("Checking if %s exists", constants.DefaultBundlePath)
	if _, err := bundle.Get(constants.GetDefaultBundle()); err != nil {
		logging.Debugf("error getting bundle info for %s: %v", constants.GetDefaultBundle(), err)
		return err
	}
	logging.Debugf("%s exists", constants.DefaultBundlePath)
	return nil
}

func fixBundleExtracted() error {
	// Should be removed after 1.19 release
	// This check will ensure correct mode for `~/.crc/cache` directory
	// in case it exists.
	if err := os.Chmod(constants.MachineCacheDir, 0775); err != nil {
		logging.Debugf("Error changing %s permissions to 0775", constants.MachineCacheDir)
	}

	if !constants.IsRelease() {
		return fmt.Errorf("CRC bundle is not embedded in the executable")
	}

	bundleDir := filepath.Dir(constants.DefaultBundlePath)
	logging.Infof("Ensuring directory %s exists", bundleDir)
	if err := os.MkdirAll(bundleDir, 0775); err != nil {
		return fmt.Errorf("Cannot create directory %s: %v", bundleDir, err)
	}

	if !crcos.FileExists(constants.DefaultBundlePath) && constants.BundleEmbedded() {
		logging.Infof("Extracting embedded bundle %s to %s", constants.GetDefaultBundle(), bundleDir)
		if err := embed.Extract(constants.GetDefaultBundle(), constants.DefaultBundlePath); err != nil {
			return err
		}
	}

	_, err := bundle.Get(constants.GetDefaultBundle())
	if err != nil {
		logging.Infof("Uncompressing %s", constants.GetDefaultBundle())
		_, err := bundle.Extract(constants.DefaultBundlePath)
		return err
	}

	return nil
}

func removeHostsFileEntry() error {
	err := adminhelper.CleanHostsFile()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func removeCRCMachinesDir() error {
	logging.Debug("Deleting machines directory")
	if err := os.RemoveAll(constants.MachineInstanceDir); err != nil {
		return fmt.Errorf("Failed to delete crc machines directory: %w", err)
	}
	return nil
}

func removeOldLogs() error {
	logFiles, err := filepath.Glob(filepath.Join(constants.CrcBaseDir, "*.log_*"))
	if err != nil {
		return fmt.Errorf("Failed to get old logs: %w", err)
	}
	for _, f := range logFiles {
		logging.Debugf("Deleting %s log file", f)
		if err := os.RemoveAll(f); err != nil {
			return fmt.Errorf("Failed to delete %s: %w", f, err)
		}
	}
	return nil
}
