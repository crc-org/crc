package cmd

import (
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/output"
	"github.com/code-ready/crc/pkg/embed"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(extractCmd)
}

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract data file embedded in the crc binary",
	Long:  `Extract a data file wich is embedded in the crc binary`,
	Run: func(cmd *cobra.Command, args []string) {
		runExtract(args)
	},
}

func runExtract(args []string) {
	if len(args) != 3 {
		logging.Fatalf("extract takes exactly three arguments")
	}
	binaryPath := args[0]
	embedName := args[1]
	destFile := args[2]
	err := embed.ExtractFromBinary(binaryPath, embedName, destFile)
	if err != nil {
		logging.Fatalf("Could not extract data embedded in %s: %v", binaryPath, err)
	}
	output.Outf("Successfully copied embedded '%s' from %s to %s: %v", embedName, binaryPath, destFile, err)
}
