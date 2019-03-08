package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start cluster",
	Long:  "Start cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runStart(args)
	},
}

func runStart(arguments []string) {
}
