package api

import (
	"encoding/json"
	"net"

	"github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/machine"
)

type newHandlerFunc func() (RequestHandler, error)
type newConfigFunc func() (config.Storage, error)
type newMachineFunc func(config.Storage) machine.Client

type commandError struct {
	Err string
}

type Server struct {
	handlerFactory         newHandlerFunc
	listener               net.Listener
	clusterOpsRequestsChan chan clusterOpsRequest
}

type RequestHandler interface {
	Start(json.RawMessage) string
	Stop() string
	Status() string
	Delete() string
	GetVersion() string
	SetConfig(json.RawMessage) string
	UnsetConfig(json.RawMessage) string
	GetConfig(json.RawMessage) string
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
