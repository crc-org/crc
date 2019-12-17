package machine

import (
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/machine/libmachine/state"
)

type GetPullSecretFunc func() (string, error)

type StartConfig struct {
	Name string

	// CRC system bundle
	BundlePath string

	// Hypervisor
	VMDriver string
	Memory   int
	CPUs     int

	// Nameserver
	NameServer string

	// Machine log output
	Debug bool

	// User Pull secret
	GetPullSecret GetPullSecretFunc
}

type ClusterConfig struct {
	KubeConfig    string
	KubeAdminPass string
	ClusterAPI    string
	WebConsoleURL string
	ProxyConfig   *network.ProxyConfig
}

type StartResult struct {
	Name           string
	Status         string
	Error          string
	ClusterConfig  ClusterConfig
	KubeletStarted bool
}

type StopConfig struct {
	Name  string
	Debug bool
}

type PowerOffConfig struct {
	Name string
}

type StopResult struct {
	Name    string
	Success bool
	State   state.State
	Error   string
}

type PowerOffResult struct {
	Name    string
	Success bool
	Error   string
}

type DeleteConfig struct {
	Name string
}

type DeleteResult struct {
	Name    string
	Success bool
	Error   string
}

type IpConfig struct {
	Name  string
	Debug bool
}

type IpResult struct {
	Name    string
	IP      string
	Success bool
	Error   string
}

type ClusterStatusConfig struct {
	Name string
}

type ClusterStatusResult struct {
	Name            string
	CrcStatus       string
	OpenshiftStatus string
	DiskUse         int64
	DiskSize        int64
	Error           string
	Success         bool
}

type ConsoleConfig struct {
	Name string
}

type ConsoleResult struct {
	ClusterConfig ClusterConfig
	State         state.State
	Success       bool
	Error         string
}
