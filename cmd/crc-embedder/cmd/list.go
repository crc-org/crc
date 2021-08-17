package cmd

import (
	"fmt"

	"github.com/YourFin/binappend"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Args:  cobra.ExactArgs(1),
	Use:   "list [path to crc executable]",
	Short: "List data files embedded in the crc executable",
	Long:  `List all the data files which were embedded in the crc executable`,
	Run: func(cmd *cobra.Command, args []string) {
		runList(args)
	},
}

func runList(args []string) {
	executablePath := args[0]
	extractor, err := binappend.MakeExtractor(executablePath)
	if err != nil {
		logging.Fatalf("Could not access data embedded in %s: %v", executablePath, err)
	}
	fmt.Printf("Data files embedded in %s:\n", executablePath)
	for _, name := range extractor.AvalibleData() {
		fmt.Println("\t", name)
	}
}
