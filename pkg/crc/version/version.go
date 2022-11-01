package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/crc-org/crc/pkg/crc/logging"
	crcPreset "github.com/crc-org/crc/pkg/crc/preset"
)

// The following variables are private fields and should be set when compiling with ldflags, for example --ldflags="-X github.com/crc-org/crc/pkg/version.crcVersion=vX.Y.Z
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
	// will always be false on linux
	// will be true for releases on macos and windows
	// will be false for git builds on macos and windows
	installerBuild = "false"
)

const (
	releaseInfoLink = "https://developers.redhat.com/content-gateway/rest/mirror/pub/openshift-v4/clients/crc/latest/release-info.json"
	// Tray version to be embedded in executable
	crcTrayElectronVersion = "1.2.8"
	crcAdminHelperVersion  = "0.0.11"
)

type CrcReleaseInfo struct {
	Version Version           `json:"version"`
	Links   map[string]string `json:"links"`
}

type Version struct {
	CrcVersion       *semver.Version `json:"crcVersion"`
	GitSha           string          `json:"gitSha"`
	OpenshiftVersion string          `json:"openshiftVersion"`
}

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
	default:
		return ocpVersion
	}
}

func GetAdminHelperVersion() string {
	return crcAdminHelperVersion
}

func GetTrayVersion() string {
	return crcTrayElectronVersion
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

func GetCRCLatestVersionFromMirror(transport http.RoundTripper) (*CrcReleaseInfo, error) {
	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}
	req, err := http.NewRequest(http.MethodGet, releaseInfoLink, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("crc/%s", crcVersion))
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %s: %d", response.Status, response.StatusCode)
	}

	releaseMetaData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var releaseInfo CrcReleaseInfo
	if err := json.Unmarshal(releaseMetaData, &releaseInfo); err != nil {
		return nil, fmt.Errorf("Error unmarshaling JSON metadata: %v", err)
	}

	return &releaseInfo, nil
}
