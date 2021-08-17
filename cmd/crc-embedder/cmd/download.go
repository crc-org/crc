package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Args:  cobra.ExactArgs(1),
	Use:   "download",
	Short: "Download data files embedded in the crc executable",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDownload(args)
	},
}

func runDownload(args []string) error {
	_, err := downloadDataFiles(runtime.GOOS, args[0])
	return err
}
