package preflight

import (
	"fmt"
	"os"
	"strings"

	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/validation"
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
		fixDescription:   "Extracting CodeReady Containers bundle",
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
}

func checkBundleExtracted() error {
	if present, _ := constants.BundlePresent(); !present {
		logging.Debug("Bundle is neither found in ~/.crc/cache nor in the same directory as the crc executable")
		return nil
	}
	if _, err := os.Stat(strings.TrimSuffix(constants.DefaultBundlePath, ".crcbundle")); os.IsNotExist(err) {
		return err
	}
	return nil
}

func fixBundleExtracted() error {
	if present, bundlePath := constants.BundlePresent(); present {
		_, err := bundle.Extract(bundlePath)
		return err
	}
	return fmt.Errorf("CRC bundle is not found in '$HOME' directories")
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
