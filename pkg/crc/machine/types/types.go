package types

import (
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/machine/state"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/preset"
	crcpreset "github.com/code-ready/crc/pkg/crc/preset"
)

type StartConfig struct {
	// CRC system bundle
	BundlePath string

	// Hypervisor
	Memory   int // Memory size in MiB
	CPUs     int
	DiskSize int // Disk size in GiB

	// Nameserver
	NameServer string

	// User Pull secret
	PullSecret cluster.PullSecretLoader

	// User defined kubeadmin password
	KubeAdminPassword string

	// Preset
	Preset crcpreset.Preset
}

type ClusterConfig struct {
	ClusterType   preset.Preset
	ClusterCACert string
	KubeConfig    string
	KubeAdminPass string
	ClusterAPI    string
	WebConsoleURL string
	ProxyConfig   *network.ProxyConfig
}

type StartResult struct {
	Status         state.State
	ClusterConfig  ClusterConfig
	KubeletStarted bool
}

type StopResult struct {
	Name    string
	Success bool
	State   state.State
	Error   string
}

type ClusterStatusResult struct {
	CrcStatus        state.State
	OpenshiftStatus  OpenshiftStatus
	OpenshiftVersion string
	PodmanVersion    string
	DiskUse          int64
	DiskSize         int64
}

type OpenshiftStatus string

const (
	OpenshiftUnreachable OpenshiftStatus = "Unreachable"
	OpenshiftStarting    OpenshiftStatus = "Starting"
	OpenshiftRunning     OpenshiftStatus = "Running"
	OpenshiftDegraded    OpenshiftStatus = "Degraded"
	OpenshiftStopped     OpenshiftStatus = "Stopped"
	OpenshiftStopping    OpenshiftStatus = "Stopping"
)

type ConsoleResult struct {
	ClusterConfig ClusterConfig
	State         state.State
}

type ConnectionDetails struct {
	IP          string
	SSHPort     int
	SSHUsername string
	SSHKeys     []string
}
