package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop cluster",
	Long:  "Stop cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runStop(args)
	},
}

func runStop(arguments []string) {
}
