package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

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

var statusFormat = `CRC VM:          {{.CrcStatus}}
OpenShift:       {{.OpenShiftStatus}}
Disk Usage:      {{.DiskUsage}}
Cache Usage:     {{.CacheUsage}}
Cache Directory: {{.CacheDir}}
`

type Status struct {
	CrcStatus       string `json:"crcStatus"`
	OpenShiftStatus string `json:"openshiftStatus"`
	DiskUsage       string `json:"diskUsage"`
	CacheUsage      string `json:"cacheUsage"`
	CacheDir        string `json:"cacheDir"`
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
	cacheUsage := units.HumanSize(float64(size))
	diskUse := units.HumanSize(float64(clusterStatus.DiskUse))
	diskSize := units.HumanSize(float64(clusterStatus.DiskSize))
	diskUsage := fmt.Sprintf("%s of %s (Inside the CRC VM)", diskUse, diskSize)
	status := Status{
		CrcStatus:       clusterStatus.CrcStatus,
		OpenShiftStatus: clusterStatus.OpenshiftStatus,
		DiskUsage:       diskUsage,
		CacheUsage:      cacheUsage,
		CacheDir:        cacheDir,
	}

	if outputFormat == jsonFormat {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(status)
	}
	return printStatus(writer, status, statusFormat)
}

func printStatus(writer io.Writer, status interface{}, statusFormat string) error {
	tmpl, err := template.New("status").Parse(statusFormat)
	if err != nil {
		return fmt.Errorf("Error creating status template: %s", err.Error())
	}
	err = tmpl.Execute(writer, status)
	if err != nil {
		return fmt.Errorf("Error executing status template: %s", err.Error())
	}
	return nil
}
