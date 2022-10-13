package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/daemonclient"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/preset"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

func init() {
	addOutputFormatFlag(statusCmd)
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of the OpenShift cluster",
	Long:  "Show details about the OpenShift cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(os.Stdout, daemonclient.New(), constants.MachineCacheDir, outputFormat)
	},
}

type status struct {
	Success          bool                         `json:"success"`
	Error            *crcErrors.SerializableError `json:"error,omitempty"`
	CrcStatus        string                       `json:"crcStatus,omitempty"`
	OpenShiftStatus  types.OpenshiftStatus        `json:"openshiftStatus,omitempty"`
	OpenShiftVersion string                       `json:"openshiftVersion,omitempty"`
	PodmanVersion    string                       `json:"podmanVersion,omitempty"`
	DiskUsage        int64                        `json:"diskUsage,omitempty"`
	DiskSize         int64                        `json:"diskSize,omitempty"`
	CacheUsage       int64                        `json:"cacheUsage,omitempty"`
	CacheDir         string                       `json:"cacheDir,omitempty"`
	RAMSize          int64                        `json:"ramSize,omitempty"`
	RAMUsage         int64                        `json:"ramUsage,omitempty"`
	Preset           preset.Preset                `json:"preset"`
}

func runStatus(writer io.Writer, client *daemonclient.Client, cacheDir, outputFormat string) error {
	status := getStatus(client, cacheDir)
	return render(status, writer, outputFormat)
}

func getStatus(client *daemonclient.Client, cacheDir string) *status {

	clusterStatus, err := client.APIClient.Status()
	if err != nil {
		var urlError *url.Error
		if errors.As(err, &urlError) {
			return &status{Success: false, Error: crcErrors.ToSerializableError(crcErrors.DaemonNotRunning)}
		}
		return &status{Success: false, Error: crcErrors.ToSerializableError(err)}
	}
	var size int64
	err = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		return &status{Success: false, Error: crcErrors.ToSerializableError(err)}
	}

	return &status{
		Success:          true,
		CrcStatus:        clusterStatus.CrcStatus,
		OpenShiftStatus:  types.OpenshiftStatus(clusterStatus.OpenshiftStatus),
		OpenShiftVersion: clusterStatus.OpenshiftVersion,
		PodmanVersion:    clusterStatus.PodmanVersion,
		DiskUsage:        clusterStatus.DiskUse,
		DiskSize:         clusterStatus.DiskSize,
		RAMSize:          clusterStatus.RAMSize,
		RAMUsage:         clusterStatus.RAMUse,
		CacheUsage:       size,
		CacheDir:         cacheDir,
		Preset:           clusterStatus.Preset,
	}
}

func (s *status) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	w := tabwriter.NewWriter(writer, 0, 0, 1, ' ', 0)

	// todo: replace this with Pair when switching to go@1.18
	type line struct {
		left, right string
	}
	lines := []line{
		{"CRC VM", s.CrcStatus},
	}

	if s.OpenShiftVersion != "" {
		lines = append(lines, line{"OpenShift", openshiftStatus(s)})
	}
	if s.PodmanVersion != "" {
		lines = append(lines, line{"Podman", s.PodmanVersion})
	}

	lines = append(lines,
		line{"RAM Usage", fmt.Sprintf(
			"%s of %s",
			units.HumanSize(float64(s.RAMUsage)),
			units.HumanSize(float64(s.RAMSize)))},
		line{"Disk Usage", fmt.Sprintf(
			"%s of %s (Inside the CRC VM)",
			units.HumanSize(float64(s.DiskUsage)),
			units.HumanSize(float64(s.DiskSize)))},
		line{"Cache Usage", units.HumanSize(float64(s.CacheUsage))},
		line{"Cache Directory", s.CacheDir})

	for _, line := range lines {
		if err := printLine(w, line.left, line.right); err != nil {
			return err
		}
	}
	return w.Flush()
}

func openshiftStatus(status *status) string {
	if status.OpenShiftVersion != "" {
		return fmt.Sprintf("%s (v%s)", status.OpenShiftStatus, status.OpenShiftVersion)
	}
	return string(status.OpenShiftStatus)
}

func printLine(w *tabwriter.Writer, left string, right string) error {
	if _, err := fmt.Fprintf(w, "%s:\t%s\n", left, right); err != nil {
		return err
	}
	return nil
}
