package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/exit"
	"github.com/code-ready/crc/pkg/crc/machine"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of the OpenShift cluster",
	Long:  "Show details about the OpenShift cluster",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

var statusFormat = `CRC VM:          {{.CrcStatus}}
OpenShift:       {{.OpenShiftStatus}}
Disk Usage:      {{.DiskUsage}}
Cache Usage:     {{.CacheUsage}}
Cache Directory: {{.CacheDir}}
`

type Status struct {
	CrcStatus       string
	OpenShiftStatus string
	DiskUsage       string
	CacheUsage      string
	CacheDir        string
}

func runStatus() {
	statusConfig := machine.ClusterStatusConfig{Name: constants.DefaultName}

	exitIfMachineMissing(statusConfig.Name)

	clusterStatus, err := machine.Status(statusConfig)
	if err != nil {
		exit.WithoutMessage(1)
	}
	cacheDir := constants.MachineCacheDir
	var size int64
	err = filepath.Walk(cacheDir, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		exit.WithMessage(1, "Error finding size of cache: %s", err.Error())
	}
	cacheUsage := units.HumanSize(float64(size))
	diskUse := units.HumanSize(float64(clusterStatus.DiskUse))
	diskSize := units.HumanSize(float64(clusterStatus.DiskSize))
	diskUsage := fmt.Sprintf("%s of %s (Inside the CRC VM)", diskUse, diskSize)
	status := Status{clusterStatus.CrcStatus, clusterStatus.OpenshiftStatus, diskUsage, cacheUsage, cacheDir}
	printStatus(status, statusFormat)
}

func printStatus(status interface{}, statusFormat string) {
	tmpl, err := template.New("status").Parse(statusFormat)
	if err != nil {
		exit.WithMessage(1, "Error creating status template: %s", err.Error())
	}
	err = tmpl.Execute(os.Stdout, status)
	if err != nil {
		exit.WithMessage(1, "Error executing status template: %s", err.Error())
	}
}
