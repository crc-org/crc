package api

import (
	"encoding/json"
	"net"

	"github.com/code-ready/crc/pkg/crc/config"

	"github.com/code-ready/crc/pkg/crc/machine"
)

type handlerFunc func(machine.Client, config.Storage, json.RawMessage) string

type commandError struct {
	Err string
}

type CrcAPIServer struct {
	client                 machine.Client
	config                 config.Storage
	listener               net.Listener
	clusterOpsRequestsChan chan clusterOpsRequest
	handlers               map[string]handlerFunc // relates commands to handler func
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
