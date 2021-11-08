package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/code-ready/crc/pkg/crc/constants"
	crcErrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/code-ready/crc/pkg/crc/machine/types"
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
		return runStatus(os.Stdout, newMachine(), constants.MachineCacheDir, outputFormat)
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
}

func runStatus(writer io.Writer, client machine.Client, cacheDir, outputFormat string) error {
	status := getStatus(client, cacheDir)
	return render(status, writer, outputFormat)
}

func getStatus(client machine.Client, cacheDir string) *status {
	if err := checkIfMachineMissing(client); err != nil {
		return &status{Success: false, Error: crcErrors.ToSerializableError(err)}
	}

	clusterStatus, err := client.Status()
	if err != nil {
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
		CrcStatus:        string(clusterStatus.CrcStatus),
		OpenShiftStatus:  clusterStatus.OpenshiftStatus,
		OpenShiftVersion: clusterStatus.OpenshiftVersion,
		PodmanVersion:    clusterStatus.PodmanVersion,
		DiskUsage:        clusterStatus.DiskUse,
		DiskSize:         clusterStatus.DiskSize,
		CacheUsage:       size,
		CacheDir:         cacheDir,
	}
}

func (s *status) prettyPrintTo(writer io.Writer) error {
	if s.Error != nil {
		return s.Error
	}
	w := tabwriter.NewWriter(writer, 0, 0, 1, ' ', 0)

	lines := []struct {
		left, right string
	}{
		{"CRC VM", s.CrcStatus},
		{"OpenShift", openshiftStatus(s)},
		{"Podman", s.PodmanVersion},
		{"Disk Usage", fmt.Sprintf(
			"%s of %s (Inside the CRC VM)",
			units.HumanSize(float64(s.DiskUsage)),
			units.HumanSize(float64(s.DiskSize)))},
		{"Cache Usage", units.HumanSize(float64(s.CacheUsage))},
		{"Cache Directory", s.CacheDir},
	}
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
