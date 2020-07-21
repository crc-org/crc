package api

import (
	"net"
)

type ArgsType map[string]string
type handlerFunc func(ArgsType) string

type commandError struct {
	Err string
}

type CrcAPIServer struct {
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
	Command string            `json:"command"`
	Args    map[string]string `json:"args,omitempty"`
}
