package cmd

import (
	"fmt"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/embed"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(extractCmd)
}

var extractCmd = &cobra.Command{
	Args:  cobra.ExactArgs(3),
	Use:   "extract [crc executable] [key identifying the embedded data] [destination filename]",
	Short: "Extract data file embedded in the crc executable",
	Long:  `Extract a data file which is embedded in the crc executable`,
	Run: func(_ *cobra.Command, args []string) {
		runExtract(args)
	},
}

func runExtract(args []string) {
	executablePath := args[0]
	embedName := args[1]
	destFile := args[2]
	err := embed.ExtractFromExecutable(executablePath, embedName, destFile)
	if err != nil {
		logging.Fatalf("Could not extract data embedded in %s: %v", executablePath, err)
	}
	fmt.Printf("Successfully copied embedded '%s' from %s to %s: %v", embedName, executablePath, destFile, err)
}
