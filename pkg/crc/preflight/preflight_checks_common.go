package preflight

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/adminhelper"
	"github.com/crc-org/crc/v2/pkg/crc/cluster"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/crc-org/crc/v2/pkg/crc/validation"
	crcos "github.com/crc-org/crc/v2/pkg/os"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
)

func bundleCheck(bundlePath string, preset crcpreset.Preset, enableBundleQuayFallback bool) Check {
	return Check{
		configKeySuffix:  "check-bundle-extracted",
		checkDescription: "Checking if CRC bundle is extracted in '$HOME/.crc'",
		check:            checkBundleExtracted(bundlePath),
		fixDescription:   "Getting bundle for the CRC executable",
		fix:              fixBundleExtracted(bundlePath, preset, enableBundleQuayFallback),
		flags:            SetupOnly,

		labels: None,
	}
}

func memoryCheck(preset crcpreset.Preset) Check {
	return Check{
		configKeySuffix:  "check-ram",
		checkDescription: "Checking minimum RAM requirements",
		check: func() error {
			return validation.ValidateEnoughMemory(constants.GetDefaultMemory(preset))
		},
		fixDescription: fmt.Sprintf("crc requires at least %s to run", units.HumanSize(float64(constants.GetDefaultMemory(preset)*1024*1024))),
		flags:          NoFix,

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
		cleanup:            removeAllLogs,
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
	{
		cleanupDescription: "Removing CRC Specific entries from user's known_hosts file",
		cleanup:            removeCRCHostEntriesFromKnownHosts,
		flags:              CleanUpOnly,

		labels: None,
	},
}

func checkBundleExtracted(bundlePath string) func() error {
	return func() error {
		logging.Infof("Checking if %s exists", bundlePath)
		bundleName := bundle.GetBundleNameFromURI(bundlePath)
		if _, err := bundle.Get(bundleName); err != nil {
			logging.Debugf("error getting bundle info for %s: %v", bundleName, err)
			return err
		}
		logging.Debugf("%s exists", bundlePath)
		return nil
	}
}

func fixBundleExtracted(bundlePath string, preset crcpreset.Preset, enableBundleQuayFallback bool) func() error {
	// Should be removed after 1.19 release
	// This check will ensure correct mode for `~/.crc/cache` directory
	// in case it exists.
	if err := os.Chmod(constants.MachineCacheDir, 0775); err != nil {
		logging.Debugf("Error changing %s permissions to 0775", constants.MachineCacheDir)
	}

	return func() error {
		bundleDir := filepath.Dir(constants.GetDefaultBundlePath(preset))
		logging.Debugf("Ensuring directory %s exists", bundleDir)
		if err := os.MkdirAll(bundleDir, 0775); err != nil {
			return fmt.Errorf("Cannot create directory %s: %v", bundleDir, err)
		}
		var err error
		logging.Infof("Downloading bundle: %s...", bundlePath)
		if bundlePath, err = bundle.Download(preset, bundlePath, enableBundleQuayFallback); err != nil {
			return err
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

func removeAllLogs() error {
	// remove all log files, need close the logfile before deletion
	logging.CloseLogging()
	if err := crcos.RemoveFileGlob(filepath.Join(constants.CrcBaseDir, "*.log")); err != nil {
		return fmt.Errorf("Failed to remove old log files: %w", err)
	}
	return nil
}

func removeCRCHostEntriesFromKnownHosts() error {
	return ssh.RemoveCRCHostEntriesFromKnownHosts()
}

func checkPodmanInOcBinDir() error {
	podmanBinPath := filepath.Join(constants.CrcOcBinDir, constants.PodmanRemoteExecutableName)
	if crcos.FileExists(podmanBinPath) {
		return fmt.Errorf("Found podman executable: %s", podmanBinPath)
	}
	return nil
}

func fixPodmanInOcBinDir() error {
	podmanBinPath := filepath.Join(constants.CrcOcBinDir, constants.PodmanRemoteExecutableName)
	if crcos.FileExists(podmanBinPath) {
		logging.Debugf("Removing podman binary at: %s", podmanBinPath)
		return os.Remove(podmanBinPath)
	}
	return nil
}

func removePodmanFromOcBinDirCheck() Check {
	return Check{
		configKeySuffix:  "check-podman-in-ocbindir",
		checkDescription: fmt.Sprintf("Check if Podman binary exists in: %s", constants.CrcOcBinDir),
		check:            checkPodmanInOcBinDir,
		fixDescription:   fmt.Sprintf("Removing Podman binary from: %s", constants.CrcOcBinDir),
		fix:              fixPodmanInOcBinDir,

		labels: None,
	}
}
