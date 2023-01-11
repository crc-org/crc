package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	downloadCmd.Flags().StringVar(&goos, "goos", runtime.GOOS, "Target platform (darwin, linux or windows)")
	downloadCmd.Flags().StringSliceVar(&components, "components", []string{}, fmt.Sprintf("List of component(s) to download (%s)", strings.Join(getAllComponentNames(goos), ", ")))
	rootCmd.AddCommand(downloadCmd)
}

var downloadCmd = &cobra.Command{
	Args:  cobra.ExactArgs(1),
	Use:   "download [destination directory]",
	Short: "Download data files to embed in the crc executable",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDownload(args)
	},
}

func runDownload(args []string) error {
	_, err := downloadDataFiles(goos, components, args[0])
	return err
}
