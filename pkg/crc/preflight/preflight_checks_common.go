package preflight

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/code-ready/crc/pkg/embed"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
)

var genericPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-podman-cached",
		checkDescription: "Checking if podman remote executable is cached",
		check:            checkPodmanExecutableCached,
		fixDescription:   "Caching podman remote executable",
		fix:              fixPodmanExecutableCached,
	},
	{
		configKeySuffix:  "check-admin-helper-cached",
		checkDescription: "Checking if admin-helper executable is cached",
		check:            checkAdminHelperExecutableCached,
		fixDescription:   "Caching admin-helper executable",
		fix:              fixAdminHelperExecutableCached,
	},
	{
		configKeySuffix:  "check-bundle-extracted",
		checkDescription: "Checking if CRC bundle is extracted in '$HOME/.crc'",
		check:            checkBundleExtracted,
		fixDescription:   "Extracting bundle from the CRC executable",
		fix:              fixBundleExtracted,
		flags:            SetupOnly,
	},
	{
		configKeySuffix:  "check-ram",
		checkDescription: "Checking minimum RAM requirements",
		check: func() error {
			return validation.ValidateEnoughMemory(constants.DefaultMemory)
		},
		fixDescription: fmt.Sprintf("crc requires at least %s to run", units.HumanSize(float64(constants.DefaultMemory*1024*1024))),
		flags:          NoFix,
	},
	{
		cleanupDescription: "Removing CRC Machine Instance directory",
		cleanup:            removeCRCMachinesDir,
		flags:              CleanUpOnly,
	},
	{
		cleanupDescription: "Removing older logs",
		cleanup:            removeOldLogs,
		flags:              CleanUpOnly,
	},
}

func checkBundleExtracted() error {
	if !constants.IsRelease() {
		return nil
	}
	if _, err := os.Stat(constants.DefaultBundlePath); os.IsNotExist(err) {
		return err
	}
	return nil
}

func fixBundleExtracted() error {
	// Should be removed after 1.19 release
	// This check will ensure correct mode for `~/.crc/cache` directory
	// in case it exists.
	if err := os.Chmod(constants.MachineCacheDir, 0775); err != nil {
		logging.Debugf("Error changing %s permissions to 0775", constants.MachineCacheDir)
	}
	if constants.IsRelease() {
		bundleDir := filepath.Dir(constants.DefaultBundlePath)
		if err := os.MkdirAll(bundleDir, 0775); err != nil {
			return fmt.Errorf("Cannot create directory %s: %v", bundleDir, err)
		}

		if err := embed.Extract(filepath.Base(constants.DefaultBundlePath), constants.DefaultBundlePath); err != nil {
			return err
		}
		_, err := bundle.Extract(constants.DefaultBundlePath)
		return err
	}
	return fmt.Errorf("CRC bundle is not embedded in the executable")
}

// Check if podman executable is cached or not
func checkPodmanExecutableCached() error {
	// Disable the podman cache until further notice
	logging.Debug("Currently podman remote is not supported")
	return nil
}

func fixPodmanExecutableCached() error {
	podman := cache.NewPodmanCache()
	if err := podman.EnsureIsCached(); err != nil {
		return fmt.Errorf("Unable to download podman remote executable %v", err)
	}
	logging.Debug("podman remote executable cached")
	return nil
}

// Check if helper executable is cached or not
func checkAdminHelperExecutableCached() error {
	helper := cache.NewAdminHelperCache()
	if !helper.IsCached() {
		return errors.New("admin-helper executable is not cached")
	}
	if err := helper.CheckVersion(); err != nil {
		return errors.Wrap(err, "unexpected version of the admin-helper executable")
	}
	logging.Debug("admin-helper executable already cached")
	return checkSuid(helper.GetExecutablePath())
}

func fixAdminHelperExecutableCached() error {
	helper := cache.NewAdminHelperCache()
	if err := helper.EnsureIsCached(); err != nil {
		return errors.Wrap(err, "Unable to download admin-helper executable")
	}
	logging.Debug("admin-helper executable cached")
	return setSuid(helper.GetExecutablePath())
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
