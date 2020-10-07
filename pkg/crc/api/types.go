package api

import (
	"encoding/json"
	"net"

	"github.com/code-ready/crc/pkg/crc/config"
)

type newConfigFunc func() (config.Storage, error)

type commandError struct {
	Err string
}

type CrcAPIServer struct {
	newConfig              newConfigFunc
	listener               net.Listener
	clusterOpsRequestsChan chan clusterOpsRequest
	handler                RequestHandler
}

type RequestHandler interface {
	Start(config.Storage, json.RawMessage) string
	Stop() string
	Status() string
	Delete() string
	GetVersion() string
	SetConfig(config.Storage, json.RawMessage) string
	UnsetConfig(config.Storage, json.RawMessage) string
	GetConfig(config.Storage, json.RawMessage) string
	GetWebconsoleInfo() string
}

// clusterOpsRequest struct is used to store the command request and associated socket
type clusterOpsRequest struct {
	command commandRequest
	socket  net.Conn
}

// commandRequest struct is used to decode the json request from tray
type commandRequest struct {
	Command string          `json:"command"`
	Args    json.RawMessage `json:"args,omitempty"`
}

// setOrUnsetConfigResult struct is used to return the result of
// setconfig/unsetconfig command
type setOrUnsetConfigResult struct {
	Error      string
	Properties []string
}

// getConfigResult struct is used to return the result of getconfig command
type getConfigResult struct {
	Error   string
	Configs map[string]interface{}
}

// startArgs is used to get the pull secret file path as argument for start handler
type startArgs struct {
	PullSecretFile string `json:"pullSecretFile"`
}
