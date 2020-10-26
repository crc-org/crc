package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	crcversion "github.com/code-ready/crc/pkg/crc/version"
	"github.com/spf13/cobra"
)

func init() {
	addOutputFormatFlag(versionCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPrintVersion(os.Stdout, defaultVersion(), outputFormat); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

func runPrintVersion(writer io.Writer, version *version, outputFormat string) error {
	return render(version, writer, outputFormat)
}

type version struct {
	Version          string `json:"version"`
	Commit           string `json:"commit"`
	OpenshiftVersion string `json:"openshiftVersion"`
	Embedded         bool   `json:"embedded"`
}

func defaultVersion() *version {
	return &version{
		Version:          crcversion.GetCRCVersion(),
		Commit:           crcversion.GetCommitSha(),
		OpenshiftVersion: crcversion.GetBundleVersion(),
		Embedded:         constants.BundleEmbedded(),
	}
}

func (v *version) prettyPrintTo(writer io.Writer) error {
	for _, line := range v.lines() {
		if _, err := fmt.Fprint(writer, line); err != nil {
			return err
		}
	}
	return nil
}

func (v *version) lines() []string {
	var embedded string
	if !v.Embedded {
		embedded = "not "
	}
	return []string{
		fmt.Sprintf("CodeReady Containers version: %s+%s\n", v.Version, v.Commit),
		fmt.Sprintf("OpenShift version: %s (%sembedded in executable)\n", v.OpenshiftVersion, embedded),
	}
}
