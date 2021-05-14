package api

import (
	"encoding/json"
	"net"
)

type Server struct {
	handler  *Handler
	listener net.Listener
}

// commandRequest struct is used to decode the json request from tray
type commandRequest struct {
	Command string          `json:"command"`
	Args    json.RawMessage `json:"args,omitempty"`
}

type loggerResult struct {
	Success  bool
	Messages []string
}
