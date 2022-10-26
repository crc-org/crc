package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
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
		return runPrintVersion(os.Stdout, getVersion(daemonclient.New()), outputFormat)
	},
}

func runPrintVersion(writer io.Writer, version *version, outputFormat string) error {
	if err := checkIfNewVersionAvailable(config.Get(crcConfig.DisableUpdateCheck).AsBool()); err != nil {
		logging.Debugf("Unable to find out if a new version is available: %v", err)
	}
	return render(version, writer, outputFormat)
}

type version struct {
	Version          string                       `json:"version"`
	Commit           string                       `json:"commit"`
	OpenshiftVersion string                       `json:"openshiftVersion"`
	PodmanVersion    string                       `json:"podmanVersion"`
	Error            *crcErrors.SerializableError `json:"error,omitempty"`
}

func defaultVersion(preset crcPreset.Preset) *version {
	return &version{
		Version:          crcversion.GetCRCVersion(),
		Commit:           crcversion.GetCommitSha(),
		OpenshiftVersion: crcversion.GetBundleVersion(preset),
		PodmanVersion:    crcversion.GetBundleVersion(crcPreset.Podman),
	}
}

func getVersion(client *daemonclient.Client) *version {
	res, err := client.APIClient.Version()
	if err != nil {
		var urlError *url.Error
		if errors.As(err, &urlError) {
			err = crcErrors.ToSerializableError(crcErrors.DaemonNotRunning)
		}
	}
	return &version{
		Version:          res.CrcVersion,
		Commit:           res.CommitSha,
		OpenshiftVersion: res.OpenshiftVersion,
		PodmanVersion:    res.PodmanVersion,
		Error:            crcErrors.ToSerializableError(err),
	}
}

func (v *version) prettyPrintTo(writer io.Writer) error {
	if v.Error != nil {
		return v.Error
	}
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
