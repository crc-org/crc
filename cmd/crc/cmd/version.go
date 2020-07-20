package cmd

import (
	"fmt"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		runPrintVersion(args)
	},
}

func GetVersionStrings() []string {
	var embedded string
	if !constants.BundleEmbedded() {
		embedded = "not "
	}
	return []string{
		fmt.Sprintf("CodeReady Containers version: %s+%s", version.GetCRCVersion(), version.GetCommitSha()),
		fmt.Sprintf("OpenShift version: %s (%sembedded in binary)", version.GetBundleVersion(), embedded),
	}
}

func runPrintVersion(arguments []string) {
	fmt.Println(strings.Join(GetVersionStrings(), "\n"))
}
