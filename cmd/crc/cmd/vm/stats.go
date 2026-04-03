package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/config"
	"github.com/crc-org/crc/v2/pkg/crc/constants"
	crcErr "github.com/crc-org/crc/v2/pkg/crc/errors"
	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/crc-org/crc/v2/pkg/crc/ssh"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"
)

func getStatsCmd(cfg *config.Config) *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Display detailed virtual machine statistics",
		Long:  "Display detailed statistics about the CRC virtual machine including OS, CPU, memory, disk, network, containers, and service health.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runStats(cfg, outputFormat)
		},
	}
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. One of: json")
	return cmd
}

type vmStats struct {
	OS         osInfo        `json:"os"`
	CPU        cpuInfo       `json:"cpu"`
	Memory     memoryInfo    `json:"memory"`
	Disk       diskInfo      `json:"disk"`
	Network    networkInfo   `json:"network"`
	Containers containerInfo `json:"containers"`
	Services   []serviceInfo `json:"services"`
	Health     healthInfo    `json:"health"`
}

type osInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kernel  string `json:"kernel"`
	Arch    string `json:"arch"`
	Uptime  int64  `json:"uptimeSeconds"`
}

type cpuInfo struct {
	Count  int     `json:"count"`
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

type memoryInfo struct {
	TotalBytes     int64   `json:"totalBytes"`
	UsedBytes      int64   `json:"usedBytes"`
	AvailableBytes int64   `json:"availableBytes"`
	UsedPercent    float64 `json:"usedPercent"`
	SwapTotalBytes int64   `json:"swapTotalBytes"`
	SwapUsedBytes  int64   `json:"swapUsedBytes"`
}

type diskInfo struct {
	TotalBytes  int64   `json:"totalBytes"`
	UsedBytes   int64   `json:"usedBytes"`
	FreeBytes   int64   `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

type networkInfo struct {
	NodeIP    string `json:"nodeIP"`
	ClusterIP string `json:"clusterIP"`
	DNSServer string `json:"dnsServer"`
}

type containerInfo struct {
	Pods       int             `json:"pods"`
	Containers int             `json:"containers"`
	Images     int             `json:"images"`
	Top        []containerStat `json:"top"`
}

type containerStat struct {
	Name     string `json:"name"`
	MemBytes uint64 `json:"memoryBytes"`
}

type serviceInfo struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type healthInfo struct {
	OOMKills        int   `json:"oomKills"`
	MajorPageFaults int64 `json:"majorPageFaults"`
}

// crictl stats JSON structures
type crictlStatsResponse struct {
	Stats []crictlContainerStats `json:"stats"`
}

type crictlContainerStats struct {
	Attributes crictlAttributes `json:"attributes"`
	Memory     crictlMemory     `json:"memory"`
}

type crictlAttributes struct {
	Labels map[string]string `json:"labels"`
}

type crictlMemory struct {
	WorkingSetBytes crictlUint64Value `json:"workingSetBytes"`
}

type crictlUint64Value struct {
	Value uint64 `json:"value,string"`
}

func runStats(cfg *config.Config, outputFormat string) error {
	client := machine.NewSynchronizedMachine(machine.NewClient(constants.DefaultName, logging.IsDebug(), cfg))
	exists, err := client.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return crcErr.VMNotExist
	}

	details, err := client.ConnectionDetails()
	if err != nil {
		return err
	}

	runner, err := ssh.CreateRunner(details.IP, details.SSHPort, details.SSHKeys...)
	if err != nil {
		return fmt.Errorf("cannot create SSH connection: %w", err)
	}
	defer runner.Close()

	stats, err := collectStats(runner)
	if err != nil {
		return err
	}
	stats.Network.NodeIP = details.IP

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(stats)
	case "":
		printStats(stats)
		return nil
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
}

func collectStats(runner *ssh.Runner) (*vmStats, error) {
	script := strings.Join([]string{
		// OS info
		`echo "===OS==="`,
		`source /etc/os-release 2>/dev/null && echo "$NAME|$VERSION|$VERSION_ID"`,
		`uname -r`,
		`uname -m`,
		// CPU & load
		`echo "===CPU==="`,
		`nproc`,
		`cat /proc/loadavg`,
		// Memory
		`echo "===MEM==="`,
		`awk '/^MemTotal:/{t=$2} /^MemAvailable:/{a=$2} /^SwapTotal:/{st=$2} /^SwapFree:/{sf=$2} END{printf "%d %d %d %d\n",t,t-a,st,st-sf}' /proc/meminfo`,
		// Disk
		`echo "===DISK==="`,
		`df -B1 --output=size,used /sysroot | tail -1`,
		// Uptime
		`echo "===UPTIME==="`,
		`awk '{print int($1)}' /proc/uptime`,
		// Network
		`echo "===NET==="`,
		`ip -br addr show br-ex 2>/dev/null | awk '{print $3}' | cut -d/ -f1`,
		`awk '/^nameserver/{print $2; exit}' /etc/resolv.conf`,
		// Containers (need sudo for crictl)
		`echo "===PODS==="`,
		`sudo crictl pods -q 2>/dev/null | wc -l`,
		`sudo crictl ps -q 2>/dev/null | wc -l`,
		`sudo crictl images -q 2>/dev/null | wc -l`,
		// Top containers by memory (raw JSON, parsed in Go)
		`echo "===TOP==="`,
		`sudo crictl stats -o json 2>/dev/null`,
		// Services
		`echo "===SVC==="`,
		`for s in kubelet crio; do echo "$s|$(systemctl is-active $s 2>/dev/null)"; done`,
		// Kernel health
		`echo "===HEALTH==="`,
		`awk '/^oom_kill /{print $2}' /proc/vmstat`,
		`awk '/^pgmajfault /{print $2}' /proc/vmstat`,
	}, "; ")

	stdout, _, err := runner.Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to collect VM stats: %w", err)
	}

	return parseStats(stdout)
}

func parseStats(output string) (*vmStats, error) {
	stats := &vmStats{}
	sections := splitSections(output)

	if lines, ok := sections["OS"]; ok && len(lines) >= 3 {
		parts := strings.SplitN(lines[0], "|", 3)
		if len(parts) >= 2 {
			stats.OS.Name = parts[0]
			stats.OS.Version = parts[1]
		}
		stats.OS.Kernel = lines[1]
		stats.OS.Arch = lines[2]
	}

	if lines, ok := sections["CPU"]; ok && len(lines) >= 2 {
		stats.CPU.Count, _ = strconv.Atoi(strings.TrimSpace(lines[0]))
		fields := strings.Fields(lines[1])
		if len(fields) >= 3 {
			stats.CPU.Load1, _ = strconv.ParseFloat(fields[0], 64)
			stats.CPU.Load5, _ = strconv.ParseFloat(fields[1], 64)
			stats.CPU.Load15, _ = strconv.ParseFloat(fields[2], 64)
		}
	}

	if lines, ok := sections["MEM"]; ok && len(lines) >= 1 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 4 {
			stats.Memory.TotalBytes, _ = strconv.ParseInt(fields[0], 10, 64)
			stats.Memory.TotalBytes *= 1024
			stats.Memory.UsedBytes, _ = strconv.ParseInt(fields[1], 10, 64)
			stats.Memory.UsedBytes *= 1024
			stats.Memory.SwapTotalBytes, _ = strconv.ParseInt(fields[2], 10, 64)
			stats.Memory.SwapTotalBytes *= 1024
			stats.Memory.SwapUsedBytes, _ = strconv.ParseInt(fields[3], 10, 64)
			stats.Memory.SwapUsedBytes *= 1024
			stats.Memory.AvailableBytes = stats.Memory.TotalBytes - stats.Memory.UsedBytes
			if stats.Memory.TotalBytes > 0 {
				stats.Memory.UsedPercent = float64(stats.Memory.UsedBytes) / float64(stats.Memory.TotalBytes) * 100
			}
		}
	}

	if lines, ok := sections["DISK"]; ok && len(lines) >= 1 {
		fields := strings.Fields(lines[0])
		if len(fields) >= 2 {
			stats.Disk.TotalBytes, _ = strconv.ParseInt(fields[0], 10, 64)
			stats.Disk.UsedBytes, _ = strconv.ParseInt(fields[1], 10, 64)
			stats.Disk.FreeBytes = stats.Disk.TotalBytes - stats.Disk.UsedBytes
			if stats.Disk.TotalBytes > 0 {
				stats.Disk.UsedPercent = float64(stats.Disk.UsedBytes) / float64(stats.Disk.TotalBytes) * 100
			}
		}
	}

	if lines, ok := sections["UPTIME"]; ok && len(lines) >= 1 {
		secs, _ := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64)
		stats.OS.Uptime = secs
	}

	if lines, ok := sections["NET"]; ok && len(lines) >= 2 {
		stats.Network.ClusterIP = strings.TrimSpace(lines[0])
		stats.Network.DNSServer = strings.TrimSpace(lines[1])
	}

	if lines, ok := sections["PODS"]; ok && len(lines) >= 3 {
		stats.Containers.Pods, _ = strconv.Atoi(strings.TrimSpace(lines[0]))
		stats.Containers.Containers, _ = strconv.Atoi(strings.TrimSpace(lines[1]))
		stats.Containers.Images, _ = strconv.Atoi(strings.TrimSpace(lines[2]))
	}

	if lines, ok := sections["TOP"]; ok {
		stats.Containers.Top = parseContainerStats(strings.Join(lines, "\n"))
	}

	if lines, ok := sections["SVC"]; ok {
		for _, line := range lines {
			parts := strings.SplitN(line, "|", 2)
			if len(parts) == 2 {
				stats.Services = append(stats.Services, serviceInfo{
					Name:   parts[0],
					Active: parts[1] == "active",
				})
			}
		}
	}

	if lines, ok := sections["HEALTH"]; ok && len(lines) >= 2 {
		stats.Health.OOMKills, _ = strconv.Atoi(strings.TrimSpace(lines[0]))
		stats.Health.MajorPageFaults, _ = strconv.ParseInt(strings.TrimSpace(lines[1]), 10, 64)
	}

	return stats, nil
}

func parseContainerStats(jsonData string) []containerStat {
	var resp crictlStatsResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		return nil
	}

	sort.Slice(resp.Stats, func(i, j int) bool {
		return resp.Stats[i].Memory.WorkingSetBytes.Value > resp.Stats[j].Memory.WorkingSetBytes.Value
	})

	limit := 5
	if len(resp.Stats) < limit {
		limit = len(resp.Stats)
	}

	result := make([]containerStat, 0, limit)
	for _, cs := range resp.Stats[:limit] {
		name := cs.Attributes.Labels["io.kubernetes.container.name"]
		if name == "" {
			name = "unknown"
		}
		result = append(result, containerStat{
			Name:     name,
			MemBytes: cs.Memory.WorkingSetBytes.Value,
		})
	}
	return result
}

func splitSections(output string) map[string][]string {
	sections := make(map[string][]string)
	var currentSection string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "===") && strings.HasSuffix(line, "===") {
			currentSection = strings.Trim(line, "=")
			continue
		}
		if currentSection != "" && line != "" {
			sections[currentSection] = append(sections[currentSection], line)
		}
	}
	return sections
}

func printStats(s *vmStats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	printSection(w, "System")
	printKV(w, "OS", fmt.Sprintf("%s %s", s.OS.Name, s.OS.Version))
	printKV(w, "Kernel", fmt.Sprintf("%s (%s)", s.OS.Kernel, s.OS.Arch))
	printKV(w, "Uptime", formatDuration(time.Duration(s.OS.Uptime)*time.Second))
	fmt.Fprintln(w)
	printSection(w, "CPU")
	printKV(w, "Cores", strconv.Itoa(s.CPU.Count))
	printKV(w, "Load Avg", fmt.Sprintf("%.2f (1m)  %.2f (5m)  %.2f (15m)", s.CPU.Load1, s.CPU.Load5, s.CPU.Load15))
	fmt.Fprintln(w)
	printSection(w, "Memory")
	printKV(w, "RAM", fmt.Sprintf("%s / %s  %s",
		units.HumanSize(float64(s.Memory.UsedBytes)),
		units.HumanSize(float64(s.Memory.TotalBytes)),
		usageBar(s.Memory.UsedPercent)))
	if s.Memory.SwapTotalBytes > 0 {
		swapPct := float64(s.Memory.SwapUsedBytes) / float64(s.Memory.SwapTotalBytes) * 100
		printKV(w, "Swap", fmt.Sprintf("%s / %s  %s",
			units.HumanSize(float64(s.Memory.SwapUsedBytes)),
			units.HumanSize(float64(s.Memory.SwapTotalBytes)),
			usageBar(swapPct)))
	} else {
		printKV(w, "Swap", "disabled")
	}
	fmt.Fprintln(w)
	printSection(w, "Disk (/sysroot)")
	printKV(w, "Usage", fmt.Sprintf("%s / %s  %s",
		units.HumanSize(float64(s.Disk.UsedBytes)),
		units.HumanSize(float64(s.Disk.TotalBytes)),
		usageBar(s.Disk.UsedPercent)))
	printKV(w, "Free", units.HumanSize(float64(s.Disk.FreeBytes)))
	fmt.Fprintln(w)
	printSection(w, "Network")
	printKV(w, "Node IP", s.Network.NodeIP)
	printKV(w, "Cluster IP", s.Network.ClusterIP)
	printKV(w, "DNS", s.Network.DNSServer)
	fmt.Fprintln(w)
	printSection(w, "Workload")
	printKV(w, "Pods", strconv.Itoa(s.Containers.Pods))
	printKV(w, "Containers", strconv.Itoa(s.Containers.Containers))
	printKV(w, "Images", strconv.Itoa(s.Containers.Images))

	if len(s.Containers.Top) > 0 {
		fmt.Fprintln(w)
		printSection(w, "Top Containers (by memory)")
		for _, c := range s.Containers.Top {
			printKV(w, c.Name, units.HumanSize(float64(c.MemBytes)))
		}
	}
	fmt.Fprintln(w)
	printSection(w, "Services")
	for _, svc := range s.Services {
		marker := "ok"
		if !svc.Active {
			marker = "NOT RUNNING"
		}
		printKV(w, svc.Name, marker)
	}
	fmt.Fprintln(w)
	printSection(w, "Health")
	printKV(w, "OOM Kills", strconv.Itoa(s.Health.OOMKills))
	printKV(w, "Major Page Faults", strconv.FormatInt(s.Health.MajorPageFaults, 10))

	fmt.Fprintln(w)
	w.Flush()
}

func printSection(w *tabwriter.Writer, title string) {
	fmt.Fprintf(w, "  %s\n", title)
	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", len(title)+2))
}

func printKV(w *tabwriter.Writer, key, value string) {
	fmt.Fprintf(w, "    %s:\t%s\n", key, value)
}

func usageBar(pct float64) string {
	const width = 20
	filled := int(pct / 100 * width)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	return fmt.Sprintf("[%s%s] %.0f%%",
		strings.Repeat("#", filled),
		strings.Repeat(".", width-filled),
		pct)
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
