package preflight

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/network"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	crcpreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/os/darwin/launchd"
)

// Deprecate warning for older version of mac<13.x
var deprecationWarning = Check{
	configKeySuffix:  "check-mac-version",
	checkDescription: "Checking if running macOS version >= 13.x",
	check:            deprecationNotice,
	fixDescription:   "This version of macOS is going to be unsupported on CRC",
	flags:            NoFix,

	labels: labels{Os: Darwin},
}

// SetupHost performs the prerequisite checks and setups the host to run the cluster
var vfkitPreflightChecks = []Check{
	{
		configKeySuffix:  "check-m1-cpu",
		checkDescription: "Checking if running emulated on Apple silicon",
		check:            checkAppleSilicon,
		fixDescription:   "This version of CRC for AMD64/Intel64 CPUs is unsupported on Apple silicon",
		flags:            NoFix,

		labels: labels{Os: Darwin},
	},
	{
		configKeySuffix:  "check-vfkit-installed",
		checkDescription: "Checking if vfkit is installed",
		check:            checkVfkitInstalled,
		fixDescription:   "Setting up virtualization with vfkit",
		fix:              fixVfkitInstallation,

		labels: labels{Os: Darwin},
	},
	{
		cleanupDescription: "Stopping CRC vfkit process",
		cleanup:            killVfkitProcess,
		flags:              CleanUpOnly,

		labels: labels{Os: Darwin},
	},
}

/*
 * Following check should be removed after 2-3 releases
 * since the tray now handles the autostart and it is unaware
 * of existing launchd configs, this is necessary to prevent
 * two instances of tray to be launched and interfering
 */
var trayLaunchdCleanupChecks = []Check{
	{
		configKeySuffix:  "check-old-autostart",
		checkDescription: "Checking if old launchd config for tray and/or daemon exists",
		check: func() error {
			if launchd.PlistExists("crc.tray") || launchd.PlistExists("crc.daemon") {
				return fmt.Errorf("plist file for tray or daemon is present")
			}
			return nil
		},
		fixDescription: "Removing old launchd config for tray and/or daemon",
		fix: func() error {
			_ = launchd.RemovePlist("crc.tray")
			_ = launchd.RemovePlist("crc.daemon")
			return nil
		},

		labels: labels{Os: Darwin},
	},
}
var resolverPreflightChecks = []Check{
	{
		configKeySuffix:    "check-resolver-file-permissions",
		checkDescription:   fmt.Sprintf("Checking file permissions for %s", resolverFile),
		check:              checkResolverFilePermissions,
		fixDescription:     fmt.Sprintf("Setting file permissions for %s", resolverFile),
		fix:                fixResolverFilePermissions,
		cleanupDescription: fmt.Sprintf("Removing %s file", resolverFile),
		cleanup:            removeResolverFile,

		labels: labels{Os: Darwin, NetworkMode: System},
	},
}

var daemonLaunchdChecks = []Check{
	{
		configKeySuffix:    "check-daemon-launchd-plist",
		checkDescription:   "Checking if crc daemon plist file is present and loaded",
		check:              checkIfDaemonPlistFileExists,
		fixDescription:     "Adding crc daemon plist file and loading it",
		fix:                fixDaemonPlistFileExists,
		cleanupDescription: "Unloading and removing the daemon plist file",
		cleanup:            removeDaemonPlistFile,
	},
}

// We want all preflight checks including
// - experimental checks
// - tray checks when using an installer, regardless of tray enabled or not
// - both user and system networking checks
//
// Passing 'SystemNetworkingMode' to getPreflightChecks currently achieves this
// as there are no user networking specific checks
func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift, false)
}

func getChecks(_ network.Mode, bundlePath string, preset crcpreset.Preset, enableBundleQuayFallback bool) []Check {
	checks := []Check{}

	checks = append(checks, deprecationWarning)
	checks = append(checks, nonWinPreflightChecks...)
	checks = append(checks, genericPreflightChecks(preset)...)
	checks = append(checks, memoryCheck(preset))
	checks = append(checks, removePodmanFromOcBinDirCheck())
	checks = append(checks, genericCleanupChecks...)
	checks = append(checks, vfkitPreflightChecks...)
	checks = append(checks, resolverPreflightChecks...)
	checks = append(checks, bundleCheck(bundlePath, preset, enableBundleQuayFallback))
	checks = append(checks, trayLaunchdCleanupChecks...)
	checks = append(checks, daemonLaunchdChecks...)
	checks = append(checks, sshPortCheck())

	return checks
}

func getPreflightChecks(_ bool, mode network.Mode, bundlePath string, preset crcpreset.Preset, enableBundleQuayFallback bool) []Check {
	filter := newFilter()
	filter.SetNetworkMode(mode)

	return filter.Apply(getChecks(mode, bundlePath, preset, enableBundleQuayFallback))
}
