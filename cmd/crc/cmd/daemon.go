package cmd

import (
	"github.com/code-ready/crc/pkg/crc/api"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cobra"
	"os"
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
		runDaemon()
	},
}

func runDaemon() {
	// Remove if an old socket is present
	os.Remove(constants.DaemonSocketPath)
	crcApiServer, err := api.CreateApiServer(constants.DaemonSocketPath)
	if err != nil {
		logging.Fatal("Failed to launch daemon", err)
	}
	crcApiServer.Serve()
}
