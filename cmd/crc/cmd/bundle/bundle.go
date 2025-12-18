package bundle

import (
	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/spf13/cobra"
)

func GetBundleCmd(config *config.Config) *cobra.Command {
	bundleCmd := &cobra.Command{
		Use:   "bundle SUBCOMMAND [flags]",
		Short: "Manage CRC bundles",
		Long:  "Manage CRC bundles, including downloading, listing, and cleaning up cached bundles.",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}
	bundleCmd.AddCommand(getGenerateCmd(config))
	bundleCmd.AddCommand(getDownloadCmd(config))
	bundleCmd.AddCommand(getListCmd(config))
	bundleCmd.AddCommand(getClearCmd())
	bundleCmd.AddCommand(getPruneCmd())
	return bundleCmd
}
