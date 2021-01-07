package cmd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "crc-embedder [list|embed|extract]",
	Short: "Build helper for crc for binary embedding",
	Long: `crc-embedder is a command line utility for listing or appending binary data
when building the crc executable for release`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		runPrerun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot()
		_ = cmd.Help()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		runPostrun()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logging.LogLevel, "log-level", constants.DefaultLogLevel, "log level (e.g. \"debug | info | warn | error\")")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Fatal(err)
	}
}

func runPrerun() {
	// Setting up logrus
	logging.InitLogrus(logging.LogLevel, constants.LogFilePath)
}

func runRoot() {
	fmt.Println("No command given")
	fmt.Println("")
}

func runPostrun() {
	logging.CloseLogging()
}
