package cmd

import (
	"fmt"
	"io"
	"os"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
	crcPreset "github.com/code-ready/crc/pkg/crc/preset"
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
		return runPrintVersion(os.Stdout, defaultVersion(crcConfig.GetPreset(config)), outputFormat)
	},
}

func runPrintVersion(writer io.Writer, version *version, outputFormat string) error {
	if err := checkIfNewVersionAvailable(config.Get(crcConfig.DisableUpdateCheck).AsBool()); err != nil {
		logging.Debugf("Unable to find out if a new version is available: %v", err)
	}
	return render(version, writer, outputFormat)
}

type version struct {
	Version          string `json:"version"`
	Commit           string `json:"commit"`
	OpenshiftVersion string `json:"openshiftVersion"`
	PodmanVersion    string `json:"podmanVersion"`
}

func defaultVersion(preset crcPreset.Preset) *version {
	return &version{
		Version:          crcversion.GetCRCVersion(),
		Commit:           crcversion.GetCommitSha(),
		OpenshiftVersion: crcversion.GetBundleVersion(preset),
		PodmanVersion:    crcversion.GetPodmanVersion(),
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
	return []string{
		fmt.Sprintf("CRC version: %s+%s\n", v.Version, v.Commit),
		fmt.Sprintf("OpenShift version: %s\n", v.OpenshiftVersion),
		fmt.Sprintf("Podman version: %s\n", v.PodmanVersion),
	}
}
