package cmd

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/output"
)

var rootCmd = &cobra.Command{
	Use:   commandName,
	Short: descriptionShort,
	Long:  descriptionLong,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		runPrerun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		runRoot()
	},
}

func init() {
	// nothing for now
}

func runPrerun() {
	output.Out("%s - %s", commandName, descriptionShort)
}

func runRoot() {
	output.Out("No command given")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logging.Log("ERR: %s", err.Error())
		os.Exit(1)
	}
}
