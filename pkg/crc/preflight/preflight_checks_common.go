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
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/pkg/errors"
)

func bundleCheck(bundlePath, preset string) Check {
	return Check{
		configKeySuffix:  "check-bundle-extracted",
		checkDescription: "Checking if CRC bundle is extracted in '$HOME/.crc'",
		check:            checkBundleExtracted(bundlePath, preset),
		fixDescription:   "Getting bundle for the CRC executable",
		fix:              fixBundleExtracted(bundlePath, preset),
		flags:            SetupOnly,

		labels: None,
	}
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

func checkBundleExtracted(bundlePath, preset string) func() error {
	return func() error {
		logging.Infof("Checking if %s exists", bundlePath)
		bundleName := filepath.Base(bundlePath)
		if _, err := bundle.Get(bundleName); err != nil {
			logging.Debugf("error getting bundle info for %s: %v", bundleName, err)
			return err
		}
		logging.Debugf("%s exists", bundlePath)
		return nil
	}
}

func fixBundleExtracted(bundlePath, preset string) func() error {
	// Should be removed after 1.19 release
	// This check will ensure correct mode for `~/.crc/cache` directory
	// in case it exists.
	if err := os.Chmod(constants.MachineCacheDir, 0775); err != nil {
		logging.Debugf("Error changing %s permissions to 0775", constants.MachineCacheDir)
	}

	return func() error {
		bundleDir := filepath.Dir(constants.DefaultBundlePath)
		logging.Infof("Ensuring directory %s exists", bundleDir)
		if err := os.MkdirAll(bundleDir, 0775); err != nil {
			return fmt.Errorf("Cannot create directory %s: %v", bundleDir, err)
		}
		if err := validation.ValidateBundle(bundlePath); err != nil {
			logging.Warnf("Provided %s not supported with this release", bundlePath)
			if err := bundle.Download(); err != nil {
				return err
			}
			bundlePath = constants.DefaultBundlePath
		}

		logging.Infof("Uncompressing %s", bundlePath)
		if _, err := bundle.Extract(bundlePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return errors.Wrap(err, "Use `crc setup -b <bundle-path>`")
			}
			return err
		}
		return nil
	}
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
