package bundle

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/gpg"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/bundle"
	"github.com/crc-org/crc/v2/pkg/crc/network/httpproxy"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/download"
	"github.com/spf13/cobra"
)

func getDownloadCmd(config *crcConfig.Config) *cobra.Command {
	downloadCmd := &cobra.Command{
		Use:   "download [version] [architecture]",
		Short: "Download a specific CRC bundle",
		Long:  "Download a specific CRC bundle from the mirrors. If no version or architecture is specified, the bundle for the current CRC version will be downloaded.",
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			force, _ := cmd.Flags().GetBool("force")
			presetStr, _ := cmd.Flags().GetString("preset")

			var preset crcPreset.Preset
			if presetStr != "" {
				var err error
				preset, err = crcPreset.ParsePresetE(presetStr)
				if err != nil {
					return err
				}
			} else {
				preset = crcConfig.GetPreset(config)
			}

			return runDownload(args, preset, verbose, force)
		},
	}
	downloadCmd.Flags().BoolP("verbose", "v", false, "Show detailed download information")
	downloadCmd.Flags().BoolP("force", "f", false, "Overwrite existing bundle if present")
	downloadCmd.Flags().StringP("preset", "p", "", "Target preset (openshift, okd, microshift)")

	downloadCmd.AddCommand(getListCmd(config))
	downloadCmd.AddCommand(getClearCmd())
	downloadCmd.AddCommand(getPruneCmd())

	return downloadCmd
}

func getListCmd(config *crcConfig.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list [version]",
		Short: "List available CRC bundles",
		Long:  "List available CRC bundles from the mirrors. Optionally filter by major.minor version (e.g. 4.19).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(args, config)
		},
	}
}

func getClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear cached CRC bundles",
		Long:  "Delete all downloaded CRC bundles from the cache directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClear()
		},
	}
}

func getPruneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prune",
		Short: "Prune old CRC bundles",
		Long:  "Keep only the most recent bundles and delete older ones to save space.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default keep 2 most recent
			return runPrune(2)
		},
	}
}

func runClear() error {
	cacheDir := constants.MachineCacheDir
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		logging.Infof("Cache directory %s does not exist", cacheDir)
		return nil
	}

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	cleared := false
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".crcbundle") {
			filePath := filepath.Join(cacheDir, file.Name())
			logging.Infof("Deleting %s", filePath)
			if err := os.RemoveAll(filePath); err != nil {
				logging.Errorf("Failed to remove %s: %v", filePath, err)
			}
			cleared = true
		}
	}

	if !cleared {
		logging.Infof("No bundles found in %s", cacheDir)
	} else {
		logging.Infof("Cleared cached bundles in %s", cacheDir)
	}
	return nil
}

func runPrune(keep int) error {
	cacheDir := constants.MachineCacheDir
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		logging.Infof("Cache directory %s does not exist", cacheDir)
		return nil
	}

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	var bundleFiles []os.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".crcbundle") {
			bundleFiles = append(bundleFiles, file)
		}
	}

	if len(bundleFiles) <= keep {
		logging.Infof("Nothing to prune (found %d bundles, keeping %d)", len(bundleFiles), keep)
		return nil
	}

	// Sort by modification time, newest first
	sort.Slice(bundleFiles, func(i, j int) bool {
		infoI, _ := bundleFiles[i].Info()
		infoJ, _ := bundleFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	for i := keep; i < len(bundleFiles); i++ {
		file := bundleFiles[i]
		filePath := filepath.Join(cacheDir, file.Name())
		logging.Infof("Pruning old bundle: %s", file.Name())
		if err := os.RemoveAll(filePath); err != nil {
			logging.Errorf("Failed to remove %s: %v", filePath, err)
		}
	}

	return nil
}

func runList(args []string, config *crcConfig.Config) error {
	preset := crcConfig.GetPreset(config)
	versions, err := fetchAvailableVersions(preset)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		logging.Infof("No bundles found for preset %s", preset)
		return nil
	}

	var filter *semver.Version
	if len(args) > 0 {
		v, err := semver.NewVersion(args[0] + ".0") // Treat 4.19 as 4.19.0 for partial matching
		if err == nil {
			filter = v
		} else {
			// Try parsing as full version just in case
			v, err = semver.NewVersion(args[0])
			if err == nil {
				filter = v
			}
		}
	}

	logging.Infof("Available bundles for %s:", preset)
	for _, v := range versions {
		if filter != nil {
			if v.Major() != filter.Major() || v.Minor() != filter.Minor() {
				continue
			}
		}

		cachedStr := ""
		if isBundleCached(preset, v.String()) {
			cachedStr = " (cached)"
		}
		fmt.Printf("%s%s\n", v.String(), cachedStr)
	}
	return nil
}

func isBundleCached(preset crcPreset.Preset, version string) bool {
	bundleName := constructBundleName(preset, version, runtime.GOARCH)
	bundlePath := filepath.Join(constants.MachineCacheDir, bundleName)
	if _, err := os.Stat(bundlePath); err == nil {
		return true
	}
	return false
}

func runDownload(args []string, preset crcPreset.Preset, verbose bool, force bool) error {
	// Disk space check (simple check for ~10GB free)
	// This is a basic check, more robust checking would require syscall/windows specific implementations
	// We skip this for now to avoid adding heavy OS-specific deps, assuming user manages disk space or download fails naturally.

	// If no args, use default bundle path
	if len(args) == 0 {
		defaultBundlePath := constants.GetDefaultBundlePath(preset)
		if !force {
			if _, err := os.Stat(defaultBundlePath); err == nil {
				logging.Infof("Bundle %s already exists. Use --force to overwrite.", defaultBundlePath)
				return nil
			}
		}

		if verbose {
			logging.Infof("Source: %s", constants.GetDefaultBundleDownloadURL(preset))
			logging.Infof("Destination: %s", defaultBundlePath)
		}
		// For default bundle, we use the existing logic which handles verification internally
		_, err := bundle.Download(context.Background(), preset, defaultBundlePath, false)
		return err
	}

	// If args provided, we are constructing a URL
	version := args[0]

	// Check if version is partial (Major.Minor) and resolve it if necessary
	resolvedVersion, err := resolveOpenShiftVersion(preset, version, verbose)
	if err != nil {
		logging.Warnf("Could not resolve version %s: %v. Trying with original version string.", version, err)
	} else if resolvedVersion != version {
		if verbose {
			logging.Infof("Resolved version %s to %s", version, resolvedVersion)
		}
		version = resolvedVersion
	}

	architecture := runtime.GOARCH
	if len(args) > 1 {
		architecture = args[1]
	}

	bundleName := constructBundleName(preset, version, architecture)
	bundlePath := filepath.Join(constants.MachineCacheDir, bundleName)

	if !force {
		if _, err := os.Stat(bundlePath); err == nil {
			logging.Infof("Bundle %s already exists. Use --force to overwrite.", bundleName)
			return nil
		}
	}

	// Base URL for the directory containing the bundle and signature
	baseVersionURL := fmt.Sprintf("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/%s/%s/", preset.String(), version)
	bundleURL := fmt.Sprintf("%s%s", baseVersionURL, bundleName)
	sigURL := fmt.Sprintf("%s%s", baseVersionURL, "sha256sum.txt.sig")

	logging.Infof("Downloading bundle: %s", bundleName)
	if verbose {
		logging.Infof("Source: %s", bundleURL)
		logging.Infof("Destination: %s", constants.MachineCacheDir)
	}

	// Implement verification logic
	logging.Infof("Verifying signature for %s...", version)
	sha256sum, err := getVerifiedHashForCustomVersion(sigURL, bundleName)
	if err != nil {
		// Fallback: try without .sig if .sig not found, maybe just sha256sum.txt?
		// For now, fail if signature verification fails as requested for "Safeguards"
		return fmt.Errorf("signature verification failed: %w", err)
	}

	sha256bytes, err := hex.DecodeString(sha256sum)
	if err != nil {
		return fmt.Errorf("failed to decode sha256sum: %w", err)
	}

	_, err = download.Download(context.Background(), bundleURL, bundlePath, 0664, sha256bytes)
	return err
}

func getVerifiedHashForCustomVersion(sigURL string, bundleName string) (string, error) {
	// Reuse existing verification logic from bundle package via a helper here
	// We essentially replicate getVerifiedHash but with our custom URL

	res, err := download.InMemory(sigURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch signature file: %w", err)
	}
	defer res.Close()

	signedHashes, err := io.ReadAll(res)
	if err != nil {
		return "", fmt.Errorf("failed to read signature file: %w", err)
	}

	verifiedHashes, err := gpg.GetVerifiedClearsignedMsgV3(constants.RedHatReleaseKey, string(signedHashes))
	if err != nil {
		return "", fmt.Errorf("invalid signature: %w", err)
	}

	lines := strings.Split(verifiedHashes, "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, bundleName) {
			sha256sum := strings.TrimSuffix(line, "  "+bundleName)
			return sha256sum, nil
		}
	}
	return "", fmt.Errorf("hash for %s not found in signature file", bundleName)
}

func constructBundleName(preset crcPreset.Preset, version string, architecture string) string {
	var bundleName strings.Builder
	bundleName.WriteString("crc")

	switch preset {
	case crcPreset.OKD:
		bundleName.WriteString("_okd")
	case crcPreset.Microshift:
		bundleName.WriteString("_microshift")
	}

	switch runtime.GOOS {
	case "darwin":
		bundleName.WriteString("_vfkit")
	case "linux":
		bundleName.WriteString("_libvirt")
	case "windows":
		bundleName.WriteString("_hyperv")
	}

	fmt.Fprintf(&bundleName, "_%s_%s.crcbundle", version, architecture)
	return bundleName.String()
}

func fetchAvailableVersions(preset crcPreset.Preset) ([]*semver.Version, error) {
	// Base URL for the preset (e.g., https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/openshift)
	baseURL := fmt.Sprintf("https://mirror.openshift.com/pub/openshift-v4/clients/crc/bundles/%s/", preset.String())

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

func resolveOpenShiftVersion(preset crcPreset.Preset, inputVersion string, verbose bool) (string, error) {
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

	if verbose {
		logging.Infof("Resolving latest version for %s...", inputVersion)
	}

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
