package cmd

import (
	"fmt"

	crcPkg "github.com/code-ready/crc/pkg/crc"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get CodeReady Container version.",
	Long:  "Get the CodeReady Container version currently installed.",
	Run: func(cmd *cobra.Command, args []string) {
		runPrintVersion(args)
	},
}

func runPrintVersion(arguments []string) {
	fmt.Printf("version: %s+%s\n", crcPkg.GetCRCVersion(), crcPkg.GetCommitSha())
}
