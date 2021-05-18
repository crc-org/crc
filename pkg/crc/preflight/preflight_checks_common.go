package preflight

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/adminhelper"
	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/code-ready/crc/pkg/embed"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
)

var bundleCheck = Check{
	configKeySuffix:  "check-bundle-extracted",
	checkDescription: "Checking if CRC bundle is extracted in '$HOME/.crc'",
	check:            checkBundleExtracted,
	fixDescription:   "Extracting bundle from the CRC executable",
	fix:              fixBundleExtracted,
	flags:            SetupOnly,
}

var genericPreflightChecks = [...]Check{
	{
		configKeySuffix:  "check-admin-helper-cached",
		checkDescription: "Checking if crc-admin-helper executable is cached",
		check:            checkAdminHelperExecutableCached,
		fixDescription:   "Caching crc-admin-helper executable",
		fix:              fixAdminHelperExecutableCached,
	},
	{

		configKeySuffix:  "check-supported-cpu-arch",
		checkDescription: "Checking if running on a supported CPU architecture",
		check:            checkSupportedCPUArch,
		fixDescription:   "CodeReady Containers is only supported on x86_64 hardware",
		flags:            NoFix,
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
		cleanupDescription: "Removing hosts file records added by CRC",
		cleanup:            removeHostsFileEntry,
		flags:              CleanUpOnly,
	},
	{
		cleanupDescription: "Removing older logs",
		cleanup:            removeOldLogs,
		flags:              CleanUpOnly,
	},
	{
		cleanupDescription: "Removing pull secret from the keyring",
		cleanup:            cluster.ForgetPullSecret,
		flags:              CleanUpOnly,
	},
}

func checkSupportedCPUArch() error {
	if runtime.GOARCH != "amd64" {
		logging.Debugf("GOARCH is %s", runtime.GOARCH)
		return fmt.Errorf("CodeReady Containers can only run on x86_64 CPUs")
	}
	return nil
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

// Check if helper executable is cached or not
func checkAdminHelperExecutableCached() error {
	if version.IsMacosInstallPathSet() || version.IsMsiBuild() {
		return nil
	}

	helper := cache.NewAdminHelperCache()
	if !helper.IsCached() {
		return errors.New("crc-admin-helper executable is not cached")
	}
	if err := helper.CheckVersion(); err != nil {
		return errors.Wrap(err, "unexpected version of the crc-admin-helper executable")
	}
	logging.Debug("crc-admin-helper executable already cached")
	return checkSuid(helper.GetExecutablePath())
}

func fixAdminHelperExecutableCached() error {
	if version.IsMacosInstallPathSet() || version.IsMsiBuild() {
		return nil
	}

	helper := cache.NewAdminHelperCache()
	if err := helper.EnsureIsCached(); err != nil {
		return errors.Wrap(err, "Unable to download crc-admin-helper executable")
	}
	logging.Debug("crc-admin-helper executable cached")
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

func removeHostsFileEntry() error {
	err := adminhelper.CleanHostsFile()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
