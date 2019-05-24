package machine

import "github.com/code-ready/machine/libmachine/state"

type StartConfig struct {
	Name string

	// CRC system bundle
	BundlePath string

	// Hypervisor
	VMDriver string
	Memory   int
	CPUs     int

	// Machine log output
	Debug bool
}

type ClusterConfig struct {
	KubeConfig    string
	KubeAdminPass string
	ClusterAPI    string
}

type StartResult struct {
	Name          string
	Status        string
	Error         string
	ClusterConfig ClusterConfig
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
	Name string
}

type IpResult struct {
	Name    string
	IP      string
	Success bool
	Error   string
}
