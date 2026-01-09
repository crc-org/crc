package bundle

import (
	"fmt"
	"runtime"

	"github.com/Masterminds/semver/v3"
	crcConfig "github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/spf13/cobra"
)

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

func runList(args []string, config *crcConfig.Config) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments: expected at most 1 version filter, got %d", len(args))
	}

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
		if isBundleCached(preset, v.String(), runtime.GOARCH) {
			cachedStr = " (cached)"
		}
		fmt.Printf("%s%s\n", v.String(), cachedStr)
	}
	return nil
}
