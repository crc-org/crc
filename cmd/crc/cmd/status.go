package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/cheggaaa/pb/v3"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	"github.com/crc-org/crc/v2/pkg/crc/daemonclient"
	crcErrors "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

var (
	watch bool
)

func init() {
	statusCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch mode, continuously update status with CPU load graph")
	addOutputFormatFlag(statusCmd)
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display status of the OpenShift cluster",
	Long:  "Show details about the OpenShift cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus(os.Stdout, daemonclient.New(), constants.MachineCacheDir, outputFormat, watch)
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

func runStatus(writer io.Writer, client *daemonclient.Client, cacheDir, outputFormat string, watch bool) error {
	if watch {
		return runWatchStatus(writer, client, cacheDir)
	}
	status := getStatus(client, cacheDir)
	return render(status, writer, outputFormat)
}

func runWatchStatus(writer io.Writer, client *daemonclient.Client, cacheDir string) error {

	status := getStatus(client, cacheDir)
	// do not render RAM size/use
	status.RAMSize = -1
	status.RAMUsage = -1
	renderError := render(status, writer, outputFormat)
	if renderError != nil {
		return renderError
	}

	var (
		barPool *pb.Pool
		cpuBars []*pb.ProgressBar
		ramBar  *pb.ProgressBar
	)

	var err error
	isPoolInit := false

	defer func() {
		if isPoolInit {
			err = barPool.Stop()
		}
	}()

	err = client.SSEClient.Status(func(loadResult *types.ClusterLoadResult) {
		if !isPoolInit {
			ramBar, cpuBars = createBars(loadResult.CPUUse, writer)
			barPool = pb.NewPool(append([]*pb.ProgressBar{ramBar}, cpuBars...)...)
			if startErr := barPool.Start(); startErr != nil {
				return
			}
			isPoolInit = true
		} else if len(loadResult.CPUUse) > len(cpuBars) {
			newCPUCount := len(loadResult.CPUUse) - len(cpuBars)
			oldCPUCount := len(cpuBars)
			for i := 0; i < newCPUCount; i++ {
				bar := createCPUBar(oldCPUCount+i, writer)
				barPool.Add(bar)
				cpuBars = append(cpuBars, bar)
			}
		}

		ramBar.SetTotal(loadResult.RAMSize)
		ramBar.SetCurrent(loadResult.RAMUse)
		for i, cpuLoad := range loadResult.CPUUse {
			cpuBars[i].SetCurrent(cpuLoad)
		}
	})
	return err
}

func createBars(cpuUse []int64, writer io.Writer) (ramBar *pb.ProgressBar, cpuBars []*pb.ProgressBar) {
	ramBar = pb.New(101)
	ramBar.SetWriter(writer)
	ramBar.Set(pb.Bytes, true)
	ramBar.Set(pb.Static, true)
	tmpl := `{{ red "RAM:" }} {{counters . }} {{percent .}} {{ bar . "[" "\u2588" "\u2588" " " "]"}} `
	ramBar.SetMaxWidth(151)
	ramBar.SetTemplateString(tmpl)

	for i := range cpuUse {
		bar := createCPUBar(i, writer)
		cpuBars = append(cpuBars, bar)
	}

	return ramBar, cpuBars
}

func createCPUBar(cpuNum int, writer io.Writer) *pb.ProgressBar {
	bar := pb.New(101)
	bar.SetWriter(writer)
	bar.Set(pb.Static, true)
	tmpl := fmt.Sprintf(`{{ green "CPU%d:" }} {{percent .}} {{ bar . "[" "\u2588" "\u2588" " " "]"}}`, cpuNum)
	bar.SetTemplateString(tmpl)
	bar.SetMaxWidth(150)
	return bar
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

	if s.Preset == preset.OpenShift {
		lines = append(lines, line{"OpenShift", openshiftStatus(s)})
	}
	if s.Preset == preset.Podman {
		lines = append(lines, line{"Podman", s.PodmanVersion})
	}
	if s.Preset == preset.Microshift {
		lines = append(lines, line{"MicroShift", openshiftStatus(s)})
	}

	if s.RAMSize != -1 && s.RAMUsage != -1 {
		lines = append(lines, line{"RAM Usage", fmt.Sprintf(
			"%s of %s",
			units.HumanSize(float64(s.RAMUsage)),
			units.HumanSize(float64(s.RAMSize)))})
	}

	lines = append(lines,
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
