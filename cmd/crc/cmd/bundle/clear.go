package bundle

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/spf13/cobra"
)

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
	var lastErr error
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".crcbundle") {
			filePath := filepath.Join(cacheDir, file.Name())
			logging.Infof("Deleting %s", filePath)
			if err := os.RemoveAll(filePath); err != nil {
				logging.Errorf("Failed to remove %s: %v", filePath, err)
				lastErr = err
			} else {
				cleared = true
			}
		}
	}

	if !cleared && lastErr == nil {
		logging.Infof("No bundles found in %s", cacheDir)
	} else if cleared {
		logging.Infof("Cleared cached bundles in %s", cacheDir)
	}
	return lastErr
}
