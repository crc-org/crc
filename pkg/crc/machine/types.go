package machine

import (
	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/state"
)

type GetPullSecretFunc func() (string, error)

type StartConfig struct {
	Name string

	// CRC system bundle
	BundlePath string

	// Hypervisor
	Memory int
	CPUs   int

	// Nameserver
	NameServer string

	// User Pull secret
	PullSecret *cluster.PullSecret
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
	Status         string
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
	CrcStatus        string
	OpenshiftStatus  string
	OpenshiftVersion string
	DiskUse          int64
	DiskSize         int64
}

type ConsoleResult struct {
	ClusterConfig ClusterConfig
	State         state.State
}

type VersionResult struct {
	CrcVersion       string
	CommitSha        string
	OpenshiftVersion string
	Success          bool
}
