package cmd

import (
	"fmt"

	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/crc/preflight"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cleanupCmd)
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: fmt.Sprintf("Undo config changes"),
	Long:  "Undo all the configuration changes done by 'crc setup' command",
	Run: func(cmd *cobra.Command, args []string) {
		runCleanup(args)
	},
}

func runCleanup(arguments []string) {
	preflight.CleanUpHost()
	output.Outln("Cleanup finished")
}
