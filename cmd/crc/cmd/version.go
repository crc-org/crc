package cmd

import (
	"fmt"
	"io"
	"os"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/logging"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPrintVersion(os.Stdout, defaultVersion(), outputFormat)
	},
}

func runPrintVersion(writer io.Writer, version *version, outputFormat string) error {
	if err := checkIfNewVersionAvailable(config.Get(crcConfig.DisableUpdateCheck).AsBool()); err != nil {
		logging.Debugf("Unable to find out if a new version is available: %v", err)
	}
	return render(version, writer, outputFormat)
}

type version struct {
	Version             string `json:"version"`
	Commit              string `json:"commit"`
	OpenshiftVersion    string `json:"openshiftVersion"`
	Embedded            bool   `json:"embedded"`
	InstalledBundlePath string `json:"installedBundlePath,omitempty"`
}

func defaultVersion() *version {
	var installedBundlePath string
	if crcversion.IsInstaller() {
		installedBundlePath = constants.DefaultBundlePath
	}
	return &version{
		Version:             crcversion.GetCRCVersion(),
		Commit:              crcversion.GetCommitSha(),
		OpenshiftVersion:    crcversion.GetBundleVersion(),
		Embedded:            constants.BundleEmbedded(),
		InstalledBundlePath: installedBundlePath,
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
	var bundleStatus string
	switch {
	case v.InstalledBundlePath != "":
		bundleStatus = fmt.Sprintf("bundle installed at %s", v.InstalledBundlePath)
	case v.Embedded:
		bundleStatus = "embedded in executable"
	default:
		bundleStatus = "not embedded in executable"
	}
	return []string{
		fmt.Sprintf("CodeReady Containers version: %s+%s\n", v.Version, v.Commit),
		fmt.Sprintf("OpenShift version: %s (%s)\n", v.OpenshiftVersion, bundleStatus),
	}
}
