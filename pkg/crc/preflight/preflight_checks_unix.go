//go:build !windows
// +build !windows

package preflight

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/code-ready/crc/pkg/crc/cache"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcpreset "github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/crc/validation"
	"github.com/code-ready/crc/pkg/crc/version"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
)

var nonWinPreflightChecks = []Check{
	{
		configKeySuffix:  "check-root-user",
		checkDescription: "Checking if running as non-root",
		check:            checkIfRunningAsNormalUser,
		fixDescription:   "crc should not be run as root",
		flags:            NoFix,

		// no need for an "os" label as this is only built on relevant OSes through the use of golang build tags
		labels: None,
	},
}

func genericPreflightChecks(preset crcpreset.Preset) []Check {
	return []Check{
		{
			configKeySuffix:  "check-admin-helper-cached",
			checkDescription: "Checking if crc-admin-helper executable is cached",
			check:            checkAdminHelperExecutableCached,
			fixDescription:   "Caching crc-admin-helper executable",
			fix:              fixAdminHelperExecutableCached,

			labels: None,
		},
		{
			configKeySuffix:  "check-obsolete-admin-helper",
			checkDescription: "Checking for obsolete admin-helper executable",
			check:            checkOldAdminHelperExecutableCached,
			fixDescription:   "Removing obsolete admin-helper executable",
			fix:              fixOldAdminHelperExecutableCached,
		},
		{
			configKeySuffix:  "check-supported-cpu-arch",
			checkDescription: "Checking if running on a supported CPU architecture",
			check:            checkSupportedCPUArch,
			fixDescription:   "CRC is only supported on AMD64/Intel64 hardware",
			flags:            NoFix,

			labels: None,
		},
		{
			configKeySuffix:  "check-ram",
			checkDescription: "Checking minimum RAM requirements",
			check: func() error {
				return validation.ValidateEnoughMemory(constants.GetDefaultMemory(preset))
			},
			fixDescription: fmt.Sprintf("crc requires at least %s to run", units.HumanSize(float64(constants.GetDefaultMemory(preset)*1024*1024))),
			flags:          NoFix,

			labels: None,
		},
		{
			configKeySuffix:    "check-crc-symlink",
			checkDescription:   "Checking if crc executable symlink exists",
			check:              checkCrcSymlink,
			fixDescription:     "Creating symlink for crc executable",
			fix:                fixCrcSymlink,
			cleanupDescription: "Removing crc executable symlink",
			cleanup:            removeCrcSymlink,

			labels: None,
		},
	}
}

func checkIfRunningAsNormalUser() error {
	if os.Geteuid() != 0 {
		return nil
	}
	logging.Debug("Ran as root")
	return errors.New("crc should not be run as root")
}

func setSuid(path string) error {
	logging.Debugf("Making %s suid", path)

	stdOut, stdErr, err := crcos.RunPrivileged(fmt.Sprintf("Changing ownership of %s", path), "chown", "root", path)
	if err != nil {
		return fmt.Errorf("Unable to set ownership of %s to root: %s %v: %s",
			path, stdOut, err, stdErr)
	}

	/* Can't do this before the chown as the chown will reset the suid bit */
	stdOut, stdErr, err = crcos.RunPrivileged(fmt.Sprintf("Setting suid for %s", path), "chmod", "u+s,g+x", path)
	if err != nil {
		return fmt.Errorf("Unable to set suid bit on %s: %s %v: %s", path, stdOut, err, stdErr)
	}
	return nil
}

func checkSuid(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Mode()&os.ModeSetuid == 0 {
		return fmt.Errorf("%s does not have the SUID bit set (%s)", path, fi.Mode().String())
	}
	if fi.Sys().(*syscall.Stat_t).Uid != 0 {
		return fmt.Errorf("%s is not owned by root", path)
	}

	return nil
}

// Check if helper executable is cached or not
func checkAdminHelperExecutableCached() error {
	if version.IsInstaller() {
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
	if version.IsInstaller() {
		return nil
	}

	helper := cache.NewAdminHelperCache()
	if err := helper.EnsureIsCached(); err != nil {
		return errors.Wrap(err, "Unable to download crc-admin-helper executable")
	}
	logging.Debug("crc-admin-helper executable cached")
	return setSuid(helper.GetExecutablePath())
}

var oldAdminHelpers = []string{"admin-helper-linux", "admin-helper-darwin"}

/* These 2 checks can be removed after a few releases */
func checkOldAdminHelperExecutableCached() error {
	logging.Debugf("Checking if an older admin-helper executable is installed")
	for _, oldExecutable := range oldAdminHelpers {
		oldPath := filepath.Join(constants.BinDir(), oldExecutable)
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			return fmt.Errorf("Found old admin-helper executable '%s'", oldExecutable)
		}
	}

	logging.Debugf("No older admin-helper executable found")

	return nil
}

func fixOldAdminHelperExecutableCached() error {
	logging.Debugf("Removing older admin-helper executable")
	for _, oldExecutable := range oldAdminHelpers {
		oldPath := filepath.Join(constants.BinDir(), oldExecutable)
		if err := os.Remove(oldPath); err != nil {
			if !os.IsNotExist(err) {
				logging.Debugf("Failed to remove  %s: %v", oldPath, err)
				return err
			}
		} else {
			logging.Debugf("Successfully removed %s", oldPath)
		}
	}

	return nil
}

func checkSupportedCPUArch() error {
	logging.Debugf("GOARCH is %s GOOS is %s", runtime.GOARCH, runtime.GOOS)
	// Only supported arches are amd64, and arm64 on macOS
	if runtime.GOARCH == "amd64" || (runtime.GOARCH == "arm64" && runtime.GOOS == "darwin") {
		return nil
	}
	return fmt.Errorf("CRC can only run on AMD64/Intel64 CPUs and M1 CPUs")
}

func runtimeExecutablePath() (string, error) {
	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		// os.Args[0] is not in $PATH, crc must have been started by specifying the path to its binary
		path = os.Args[0]
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

func checkCrcSymlink() error {
	runtimePath, err := runtimeExecutablePath()
	if err != nil {
		return err
	}
	symlinkPath, err := filepath.EvalSymlinks(constants.CrcSymlinkPath)
	if err != nil {
		return err
	}
	if symlinkPath != runtimePath {
		return fmt.Errorf("%s points to %s, not to %s", constants.CrcSymlinkPath, symlinkPath, runtimePath)
	}

	return nil
}

func fixCrcSymlink() error {
	_ = os.Remove(constants.CrcSymlinkPath)

	runtimePath, err := runtimeExecutablePath()
	if err != nil {
		return err
	}
	logging.Debugf("symlinking %s to %s", runtimePath, constants.CrcSymlinkPath)
	return os.Symlink(runtimePath, constants.CrcSymlinkPath)
}

func removeCrcSymlink() error {
	if crcos.FileExists(constants.CrcSymlinkPath) {
		return os.Remove(constants.CrcSymlinkPath)
	}
	return nil
}
