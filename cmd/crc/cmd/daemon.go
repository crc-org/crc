package cmd

import (
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
}

var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run the crc daemon",
	Long:   "Run the crc daemon",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		// setup separate logging for daemon
		logging.CloseLogging()
		logging.InitLogrus(logging.LogLevel, constants.DaemonLogFilePath)

		runDaemon(config)
	},
}
