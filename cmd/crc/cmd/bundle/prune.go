package bundle

import (
	"os"
	"path/filepath"
	"sort"
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
		infoI, errI := bundleFiles[i].Info()
		infoJ, errJ := bundleFiles[j].Info()
		if errI != nil || errJ != nil {
			// If we can't get info, treat as oldest (sort to end for pruning)
			return errJ != nil && errI == nil
		}
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
