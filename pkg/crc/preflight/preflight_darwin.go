package preflight

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/preset"
	crcpreset "github.com/code-ready/crc/pkg/crc/preset"
	"github.com/code-ready/crc/pkg/os/launchd"
)

// SetupHost performs the prerequisite checks and setups the host to run the cluster
func hyperkitPreflightChecks(networkMode network.Mode) []Check {
	return []Check{
		{
			configKeySuffix:  "check-m1-cpu",
			checkDescription: "Checking if running emulated on a M1 CPU",
			check:            checkM1CPU,
			fixDescription:   "CodeReady Containers is unsupported on Apple M1 hardware",
			flags:            NoFix,

			labels: labels{Os: Darwin},
		},
		{
			configKeySuffix:  "check-hyperkit-installed",
			checkDescription: "Checking if HyperKit is installed",
			check:            checkHyperKitInstalled(networkMode),
			fixDescription:   "Setting up virtualization with HyperKit",
			fix:              fixHyperKitInstallation(networkMode),

			labels: labels{Os: Darwin},
		},
		{
			configKeySuffix:  "check-qcow-tool-installed",
			checkDescription: "Checking if qcow-tool is installed",
			check:            checkQcowToolInstalled,
			fixDescription:   "Installing qcow-tool",
			fix:              fixQcowToolInstalled,

			labels: labels{Os: Darwin},
		},
		{
			configKeySuffix:  "check-hyperkit-driver",
			checkDescription: "Checking if crc-driver-hyperkit is installed",
			check:            checkMachineDriverHyperKitInstalled(networkMode),
			fixDescription:   "Installing crc-machine-hyperkit",
			fix:              fixMachineDriverHyperKitInstalled(networkMode),

			labels: labels{Os: Darwin},
		},
		{
			cleanupDescription: "Stopping CRC Hyperkit process",
			cleanup:            stopCRCHyperkitProcess,
			flags:              CleanUpOnly,

			labels: labels{Os: Darwin},
		},
	}
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
		checkDescription: "Checking if old launchd config for tray autostart exists",
		check: func() error {
			if launchd.PlistExists("crc.tray") || launchd.PlistExists("crc.daemon") {
				return fmt.Errorf("force trigger cleanup to remove old launchd config for tray")
			}
			return nil
		},
		fixDescription: "Removing old launchd config for tray autostart",
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

// We want all preflight checks including
// - experimental checks
// - tray checks when using an installer, regardless of tray enabled or not
// - both user and system networking checks
//
// Passing 'SystemNetworkingMode' to getPreflightChecks currently achieves this
// as there are no user networking specific checks
func getAllPreflightChecks() []Check {
	return getPreflightChecks(true, network.SystemNetworkingMode, constants.GetDefaultBundlePath(preset.OpenShift), preset.OpenShift)
}

func getChecks(mode network.Mode, bundlePath string, preset crcpreset.Preset) []Check {
	checks := []Check{}

	checks = append(checks, nonWinPreflightChecks...)
	checks = append(checks, genericPreflightChecks(preset)...)
	checks = append(checks, genericCleanupChecks...)
	checks = append(checks, hyperkitPreflightChecks(mode)...)
	checks = append(checks, resolverPreflightChecks...)
	checks = append(checks, bundleCheck(bundlePath, preset))
	checks = append(checks, trayLaunchdCleanupChecks...)

	return checks
}

func getPreflightChecks(_ bool, mode network.Mode, bundlePath string, preset crcpreset.Preset) []Check {
	filter := newFilter()
	filter.SetNetworkMode(mode)

	return filter.Apply(getChecks(mode, bundlePath, preset))
}
