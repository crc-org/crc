package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/spf13/cobra"
)

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

type bundleVersionInfo struct {
	name  string
	major int
	minor int
	patch int
	arch  string
}

var bundleVersionRegex = regexp.MustCompile(`^crc(?:_okd|_microshift)?_(?:vfkit|libvirt|hyperv)_(\d+)\.(\d+)\.(\d+)_([a-z0-9]+)\.crcbundle$`)

// parseBundleVersion parses a bundle filename like "crc_vfkit_4.19.13_arm64.crcbundle"
// and extracts the version parts and architecture.
func parseBundleVersion(filename string) (bundleVersionInfo, error) {
	matches := bundleVersionRegex.FindStringSubmatch(filename)
	if matches == nil {
		return bundleVersionInfo{}, fmt.Errorf("filename %q does not match expected bundle pattern", filename)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return bundleVersionInfo{}, fmt.Errorf("invalid major version in %q: %w", filename, err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return bundleVersionInfo{}, fmt.Errorf("invalid minor version in %q: %w", filename, err)
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return bundleVersionInfo{}, fmt.Errorf("invalid patch version in %q: %w", filename, err)
	}

	return bundleVersionInfo{
		name:  filename,
		major: major,
		minor: minor,
		patch: patch,
		arch:  matches[4],
	}, nil
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

	// Group bundles by major.minor + arch
	type groupKey struct {
		major int
		minor int
		arch  string
	}
	groups := make(map[groupKey][]bundleVersionInfo)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".crcbundle") {
			continue
		}
		info, err := parseBundleVersion(file.Name())
		if err != nil {
			logging.Debugf("Skipping unrecognized bundle file: %s", file.Name())
			continue
		}
		key := groupKey{major: info.major, minor: info.minor, arch: info.arch}
		groups[key] = append(groups[key], info)
	}

	pruned := false
	var lastErr error
	for _, bundles := range groups {
		if len(bundles) <= keep {
			continue
		}

		// Sort by patch version descending (newest first)
		sort.Slice(bundles, func(i, j int) bool {
			return bundles[i].patch > bundles[j].patch
		})

		for i := keep; i < len(bundles); i++ {
			filePath := filepath.Join(cacheDir, bundles[i].name)
			logging.Infof("Pruning old bundle: %s", bundles[i].name)
			if err := os.Remove(filePath); err != nil {
				logging.Errorf("failed to remove %s: %v", filePath, err)
				lastErr = err
			} else {
				pruned = true
				// Also remove extracted bundle directory if it exists
				dirPath := strings.TrimSuffix(filePath, ".crcbundle")
				if _, err := os.Stat(dirPath); err == nil {
					if err := os.RemoveAll(dirPath); err != nil {
						logging.Errorf("failed to remove extracted bundle %s: %v", dirPath, err)
						lastErr = err
					}
				}
			}
		}
	}

	if !pruned && lastErr == nil {
		logging.Infof("Nothing to prune")
	}

	return lastErr
}
