package machine

import (
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/state"
)

type StartConfig struct {
	// CRC system bundle
	BundlePath string

	// Hypervisor
	Memory   int // Memory size in MB
	CPUs     int
	DiskSize int // Disk size in GiB

	// Nameserver
	NameServer string

	// User Pull secret
	PullSecret cluster.PullSecretLoader
}

type ClusterConfig struct {
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
	OpenshiftStatus  string
	OpenshiftVersion string
	DiskUse          int64
	DiskSize         int64
}

type ConsoleResult struct {
	ClusterConfig ClusterConfig
	State         state.State
}
