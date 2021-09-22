package cmd

import (
	"fmt"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "crc-embedder [command]",
	Short: "Build helper for crc for binary embedding",
	Long: `crc-embedder is a command line utility for listing or appending binary data
when building the crc executable for release`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		runPrerun()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		runPostrun()
	},
}

func init() {
	err := constants.EnsureBaseDirectoriesExist()
	if err != nil {
		fmt.Println("CRC base directories are missing: ", err)
		os.Exit(1)
	}
	logging.AddLogLevelFlag(rootCmd.PersistentFlags())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Fatal(err)
	}
}

func runPrerun() {
	logging.InitLogrus(constants.LogFilePath)
}

func runPostrun() {
	logging.CloseLogging()
}
