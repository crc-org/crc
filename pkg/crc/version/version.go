package version

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Masterminds/semver"
)

// The following variables are private fields and should be set when compiling with ldflags, for example --ldflags="-X github.com/code-ready/crc/pkg/version.crcVersion=vX.Y.Z
var (
	// The current version of minishift
	crcVersion = "0.0.0-unset"

	// The SHA-1 of the commit this executable is build off
	commitSha = "sha-unset"

	// Bundle version which used for the release.
	bundleVersion = "0.0.0-unset"

	okdBuild = "false"

	macosInstallPath = "/unset"

	msiBuild = "false"
)

const (
	releaseInfoLink = "https://developers.redhat.com/content-gateway/file/pub/openshift-v4/clients/crc/latest/release-info.json"
	// Tray version to be embedded in executable
	crcMacTrayVersion = "1.0.8"
	// Windows forms application version type major.minor.buildnumber.revesion
	crcWindowsTrayVersion = "0.5.0.0"
)

type CrcReleaseInfo struct {
	Version struct {
		LatestVersion string `json:"crcVersion"`
	}
}

func GetCRCVersion() string {
	return crcVersion
}

func GetCommitSha() string {
	return commitSha
}

func GetBundleVersion() string {
	return bundleVersion
}

func IsOkdBuild() bool {
	return okdBuild == "true"
}

func GetCRCMacTrayVersion() string {
	return crcMacTrayVersion
}

func GetCRCWindowsTrayVersion() string {
	return crcWindowsTrayVersion
}

func GetMacosInstallPath() string {
	return macosInstallPath
}

func IsMacosInstallPathSet() bool {
	return macosInstallPath != "/unset"
}

func IsMsiBuild() bool {
	return msiBuild != "false"
}

func getCRCLatestVersionFromMirror() (*semver.Version, error) {
	var releaseInfo CrcReleaseInfo
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	response, err := client.Get(releaseInfoLink)
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
	err = json.Unmarshal(releaseMetaData, &releaseInfo)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling JSON metadata: %v", err)
	}
	version, err := semver.NewVersion(releaseInfo.Version.LatestVersion)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func NewVersionAvailable() (bool, string, error) {
	latestVersion, err := getCRCLatestVersionFromMirror()
	if err != nil {
		return false, "", err
	}
	currentVersion, err := semver.NewVersion(GetCRCVersion())
	if err != nil {
		return false, "", err
	}
	return latestVersion.GreaterThan(currentVersion), latestVersion.String(), nil
}
