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

	resp, err := client.Do(req)
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

	// Parse the HTML directory listing to find version directories
	versionRegex := regexp.MustCompile(`href=["']?\.?/?(\d+\.\d+\.\d+)/?["']?`)

	matches := versionRegex.FindAllStringSubmatch(string(body), -1)

	var versions []*semver.Version
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			vStr := match[1]
			if seen[vStr] {
				continue
			}
			v, err := semver.NewVersion(vStr)
			if err == nil {
				versions = append(versions, v)
				seen[vStr] = true
			}
		}
	}

	// If regex failed, try a simpler one for directory names in text
	if len(versions) == 0 {
		simpleRegex := regexp.MustCompile(`>(\d+\.\d+\.\d+)/?<`)
		matches = simpleRegex.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) > 1 {
				vStr := match[1]
				if seen[vStr] {
					continue
				}
				v, err := semver.NewVersion(vStr)
				if err == nil {
					versions = append(versions, v)
					seen[vStr] = true
				}
			}
		}
	}

	// Sort reverse (newest first)
	sort.Sort(sort.Reverse(semver.Collection(versions)))
	return versions, nil
}

func resolveOpenShiftVersion(preset crcPreset.Preset, inputVersion string) (string, error) {
	// If input already looks like a full version (Major.Minor.Patch), return as is
	fullVersionRegex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	if fullVersionRegex.MatchString(inputVersion) {
		return inputVersion, nil
	}

	// If not Major.Minor, return as is (could be a tag or other format user intends)
	partialVersionRegex := regexp.MustCompile(`^(\d+\.\d+)$`)
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
	bundleName := constants.BundleName(preset, version, arch)
	bundlePath := filepath.Join(constants.MachineCacheDir, bundleName)
	if _, err := os.Stat(bundlePath); err == nil {
		return true
	}
	return false
}
