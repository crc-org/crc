package image

import (
	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/spf13/cobra"
)

func GetImageCmd(config *config.Config) *cobra.Command {
	imageCmd := &cobra.Command{
		Use:   "image SUBCOMMAND [flags]",
		Short: "Manage container images in the CRC cluster",
		Long:  "Manage container images in the CRC cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}
	imageCmd.AddCommand(getLoadCmd(config))
	return imageCmd
}
