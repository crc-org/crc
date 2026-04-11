package bundle

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
)

var (
	// versionHrefPattern matches version strings in HTML href attributes.
	versionHrefPattern = regexp.MustCompile(`href=["']?\.?/?(\d+\.\d+\.\d+)/?["']?`)
	// versionTextPattern matches version strings in bare HTML text.
	versionTextPattern = regexp.MustCompile(`>(\d+\.\d+\.\d+)/?<`)
	// fullVersionRegex matches a complete Major.Minor.Patch version string.
	fullVersionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	// partialVersionRegex matches a Major.Minor version string.
	partialVersionRegex = regexp.MustCompile(`^\d+\.\d+$`)
)

func fetchAvailableVersions(preset crcPreset.Preset) ([]*semver.Version, error) {
	// Base URL for the preset
	baseURL := fmt.Sprintf("%s/%s/", constants.DefaultMirrorURL, preset.String())

	client := &http.Client{
		Transport: httpproxy.HTTPTransport(),
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req) // nolint:gosec // G704: URL is constructed from constant base URL and validated preset enum
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch versions from mirror: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the HTML directory listing to find version directories.
	// Try the href pattern first, fall back to bare text pattern.
	versionPatterns := []*regexp.Regexp{versionHrefPattern, versionTextPattern}

	var versions []*semver.Version
	seen := make(map[string]bool)

	for _, pattern := range versionPatterns {
		for _, match := range pattern.FindAllStringSubmatch(string(body), -1) {
			if len(match) > 1 && !seen[match[1]] {
				v, err := semver.NewVersion(match[1])
				if err == nil {
					versions = append(versions, v)
					seen[match[1]] = true
				}
			}
		}
		if len(versions) > 0 {
			break
		}
	}

	// Sort reverse (newest first)
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	return versions, nil
}

func resolveOpenShiftVersion(preset crcPreset.Preset, inputVersion string) (string, error) {
	// If input already looks like a full version (Major.Minor.Patch), return as is
	if fullVersionRegex.MatchString(inputVersion) {
		return inputVersion, nil
	}

	// If not Major.Minor, return as is (could be a tag or other format user intends)
	if !partialVersionRegex.MatchString(inputVersion) {
		return inputVersion, nil
	}

	logging.Debugf("Resolving latest version for %s...", inputVersion)

	versions, err := fetchAvailableVersions(preset)
	if err != nil {
		return "", err
	}

	inputVer, err := semver.NewVersion(inputVersion + ".0")
	if err != nil {
		return "", fmt.Errorf("invalid input version format: %v", err)
	}

	for _, v := range versions {
		if v.Major() == inputVer.Major() && v.Minor() == inputVer.Minor() {
			return v.String(), nil
		}
	}

	return "", fmt.Errorf("no matching versions found for %s", inputVersion)
}

func isBundleCached(preset crcPreset.Preset, version string, arch string) bool {
	bundlePath := filepath.Join(constants.MachineCacheDir, constants.BundleName(preset, version, arch))
	_, err := os.Stat(bundlePath)
	return err == nil
}
