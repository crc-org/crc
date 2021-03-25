package bundle

import (
	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/spf13/cobra"
)

func GetBundleCmd(config *config.Config) *cobra.Command {
	bundleCmd := &cobra.Command{
		Use:   "bundle SUBCOMMAND [flags]",
		Short: "Manage CRC bundles",
		Long:  "Manage CRC bundles",
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}
	bundleCmd.AddCommand(getGenerateCmd(config))
	return bundleCmd
}
