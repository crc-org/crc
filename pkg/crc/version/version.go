package version

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
)

// The following variables are private fields and should be set when compiling with ldflags, for example --ldflags="-X github.com/crc-org/crc/v2/pkg/version.crcVersion=vX.Y.Z
var (
	// The current version of minishift
	crcVersion = "0.0.0-unset"

	// The SHA-1 of the commit this executable is build off
	commitSha = "sha-unset"

	// OCP version which is used for the release.
	ocpVersion = "0.0.0-unset"

	// Podman version for podman specific bundles
	podmanVersion = "0.0.0-unset"

	okdVersion = "0.0.0-unset"

	microshiftVersion = "0.0.0-unset"
	// will always be false on linux
	// will be true for releases on macos and windows
	// will be false for git builds on macos and windows
	installerBuild = "false"

	defaultPreset = "openshift"
)

const (
	crcAdminHelperVersion          = "0.5.2"
	win32BackgroundLauncherVersion = "0.0.0.1"
)

func GetCRCVersion() string {
	return crcVersion
}

func GetCommitSha() string {
	return commitSha
}

func GetBundleVersion(preset crcPreset.Preset) string {
	switch preset {
	case crcPreset.Podman:
		return podmanVersion
	case crcPreset.OKD:
		return okdVersion
	case crcPreset.Microshift:
		return microshiftVersion
	default:
		return ocpVersion
	}
}

func GetAdminHelperVersion() string {
	return crcAdminHelperVersion
}

func GetWin32BackgroundLauncherVersion() string {
	return win32BackgroundLauncherVersion
}

func IsInstaller() bool {
	return installerBuild != "false"
}

func InstallPath() string {
	if !IsInstaller() {
		logging.Errorf("version.InstallPath() should not be called on non-installer builds")
		return ""
	}

	// In case of error, path will be an empty string and the upper layers will handle it
	currentExecutablePath, err := os.Executable()
	if err != nil {
		logging.Errorf("Failed to find the installation path: %v", err)
		return ""
	}
	src, err := filepath.EvalSymlinks(currentExecutablePath)
	if err != nil {
		logging.Errorf("Failed to find the symlink destination of %s: %v", currentExecutablePath, err)
		return filepath.Dir(currentExecutablePath)
	}
	return filepath.Dir(src)
}

func GetDefaultPreset() crcPreset.Preset {
	preset, err := crcPreset.ParsePresetE(defaultPreset)
	if err != nil {
		// defaultPreset is set at compile-time, it should *never* be invalid
		panic(fmt.Sprintf("Invalid compilet-time default preset '%s'", defaultPreset))
	}
	return preset
}

func UserAgent() string {
	return fmt.Sprintf("crc/%s", crcVersion)
}
