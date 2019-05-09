package daemon

import (
	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts a CRC service daemon.",
	Long:  `Starts a CRC service daemon. This command is for internal use of CRC services`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Hidden: true,
}
