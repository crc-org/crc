package vm

import (
	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/spf13/cobra"
)

func GetVMCmd(cfg *config.Config) *cobra.Command {
	vmCmd := &cobra.Command{
		Use:   "vm SUBCOMMAND [flags]",
		Short: "Inspect and manage the CRC virtual machine",
		Long:  "Inspect and manage the CRC virtual machine",
		Run: func(cmd *cobra.Command, _ []string) {
			_ = cmd.Help()
		},
	}
	vmCmd.AddCommand(getStatsCmd(cfg))
	return vmCmd
}
