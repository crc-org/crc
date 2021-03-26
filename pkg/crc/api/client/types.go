package client

import (
	"github.com/code-ready/crc/pkg/crc/machine/types"
)

type VersionResult struct {
	CrcVersion       string
	CommitSha        string
	OpenshiftVersion string
	Success          bool
}

type Result struct {
	Name    string
	Success bool
	Error   string
}

type StartResult struct {
	Name           string
	Status         string
	Error          string
	ClusterConfig  types.ClusterConfig
	KubeletStarted bool
}

type ClusterStatusResult struct {
	Name             string
	CrcStatus        string
	OpenshiftStatus  string
	OpenshiftVersion string
	DiskUse          int64
	DiskSize         int64
	Error            string
	Success          bool
}

type ConsoleResult struct {
	ClusterConfig types.ClusterConfig
	Success       bool
	Error         string
}

// setOrUnsetConfigResult struct is used to return the result of
// setconfig/unsetconfig command
type SetOrUnsetConfigResult struct {
	Error      string
	Properties []string
}

// getConfigResult struct is used to return the result of getconfig command
type GetConfigResult struct {
	Error   string
	Configs map[string]interface{}
}

type StartConfig struct {
	PullSecretPath string
}

type SetConfigRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

type getOrUnsetConfigRequest struct {
	Properties []string `json:"properties"`
}
