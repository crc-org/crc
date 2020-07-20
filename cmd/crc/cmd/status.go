package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

const jsonFormat = "json"

var (
	outputFormat string
)

func init() {
	statusCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. One of: json")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of the OpenShift cluster",
	Long:  "Show details about the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStatus(os.Stdout, &libmachineClient{}, constants.MachineCacheDir, outputFormat); err != nil {
			exit.WithMessage(1, err.Error())
		}
	},
}

type Status struct {
	CrcStatus        string `json:"crcStatus"`
	OpenShiftStatus  string `json:"openshiftStatus"`
	OpenShiftVersion string `json:"openshiftVersion"`
	DiskUsage        int64  `json:"diskUsage"`
	DiskSize         int64  `json:"diskSize"`
	CacheUsage       int64  `json:"cacheUsage"`
	CacheDir         string `json:"cacheDir"`
}

func runStatus(writer io.Writer, client client, cacheDir, outputFormat string) error {
	statusConfig := machine.ClusterStatusConfig{Name: constants.DefaultName}

	if err := checkIfMachineMissing(client, statusConfig.Name); err != nil {
		return err
	}

	clusterStatus, err := client.Status(statusConfig)
	if err != nil {
		return err
	}
	var size int64
	err = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("Error finding size of cache: %s", err.Error())
	}
	status := Status{
		CrcStatus:        clusterStatus.CrcStatus,
		OpenShiftStatus:  clusterStatus.OpenshiftStatus,
		OpenShiftVersion: clusterStatus.OpenshiftVersion,
		DiskUsage:        clusterStatus.DiskUse,
		DiskSize:         clusterStatus.DiskSize,
		CacheUsage:       size,
		CacheDir:         cacheDir,
	}

	if outputFormat == jsonFormat {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(status)
	}
	return printStatus(writer, status)
}

func printStatus(writer io.Writer, status Status) error {
	w := tabwriter.NewWriter(writer, 0, 0, 1, ' ', 0)

	lines := []struct {
		left, right string
	}{
		{"CRC VM", status.CrcStatus},
		{"OpenShift", openshiftStatus(status)},
		{"Disk Usage", fmt.Sprintf(
			"%s of %s (Inside the CRC VM)",
			units.HumanSize(float64(status.DiskUsage)),
			units.HumanSize(float64(status.DiskSize)))},
		{"Cache Usage", units.HumanSize(float64(status.CacheUsage))},
		{"Cache Directory", status.CacheDir},
	}
	for _, line := range lines {
		if err := printLine(w, line.left, line.right); err != nil {
			return err
		}
	}
	return w.Flush()
}

func openshiftStatus(status Status) string {
	if status.OpenShiftVersion != "" {
		return fmt.Sprintf("%s (v%s)", status.OpenShiftStatus, status.OpenShiftVersion)
	}
	return status.OpenShiftStatus
}

func printLine(w *tabwriter.Writer, left string, right string) error {
	if _, err := fmt.Fprintf(w, "%s:\t%s\n", left, right); err != nil {
		return err
	}
	return nil
}
