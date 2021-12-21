package client

import (
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/preset"
)

type VersionResult struct {
	CrcVersion       string
	CommitSha        string
	OpenshiftVersion string
	Success          bool
}

type Result struct {
	Success bool
	Error   string
}

type StartResult struct {
	Success        bool
	Status         string
	Error          string
	ClusterConfig  types.ClusterConfig
	KubeletStarted bool
}

type ClusterStatusResult struct {
	CrcStatus        string
	OpenshiftStatus  string
	OpenshiftVersion string
	PodmanVersion    string
	DiskUse          int64
	DiskSize         int64
	Error            string
	Success          bool
	Preset           preset.Preset
}

type ConsoleResult struct {
	ClusterConfig types.ClusterConfig
	Success       bool
	Error         string
}

// setOrUnsetConfigResult struct is used to return the result of
// setconfig/unsetconfig command
type SetOrUnsetConfigResult struct {
	Success    bool
	Error      string
	Properties []string
}

// getConfigResult struct is used to return the result of getconfig command
type GetConfigResult struct {
	Success bool
	Error   string
	Configs map[string]interface{}
}

type StartConfig struct {
	PullSecretFile string `json:"pullSecretFile"`
}

type SetConfigRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

type GetOrUnsetConfigRequest struct {
	Properties []string `json:"properties"`
}

type TelemetryRequest struct {
	Action string `json:"action"`
	Source string `json:"source"`
	Status string `json:"status"`
}
