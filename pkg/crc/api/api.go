package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"github.com/code-ready/crc/pkg/crc/logging"
)

type ArgsType map[string]string
type handlerFunc func(ArgsType) string

type CrcApiServer struct {
	listener net.Listener
	handlers map[string]handlerFunc // relates commands to handler func
}

// commandRequest struct is used to decode the json request from tray
type commandRequest struct {
	Command string            `json:"command"`
	Args    map[string]string `json:"args,omitempty"`
}

func CreateApiServer(socketPath string) (CrcApiServer, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logging.Error("Failed to create socket: ", err.Error())
		return CrcApiServer{}, err
	}
	apiServer := CrcApiServer{
		listener: listener,
		handlers: map[string]handlerFunc{
			"start":         startHandler,
			"stop":          stopHandler,
			"status":        statusHandler,
			"delete":        deleteHandler,
			"version":       versionHandler,
			"webconsoleurl": webconsoleURLHandler,
		},
	}
	return apiServer, nil
}

func (api CrcApiServer) Serve() {
	for {
		conn, err := api.listener.Accept()
		if err != nil {
			logging.Error("Error establishing communication: ", err.Error())
			continue
		}
		go api.handleConnection(conn)
	}
}

func (api CrcApiServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	inBuffer := make([]byte, 1024)
	var req commandRequest
	numBytes, err := conn.Read(inBuffer)
	if err != nil || numBytes == 0 || numBytes == cap(inBuffer) {
		logging.Error("Error reading from socket")
		return
	}
	logging.Debug("Received Request:", string(inBuffer[0:numBytes]))
	err = json.Unmarshal(inBuffer[0:numBytes], &req)
	if err != nil {
		logging.Error("Error decoding request: ", err.Error())
		return
	}

	if handler, ok := api.handlers[req.Command]; ok {
		result := handler(req.Args)
		writeStringToSocket(conn, result)
	} else {
		writeStringToSocket(conn, fmt.Sprintf("Unknown command supplied: %s", req.Command))
	}
}

func writeStringToSocket(socket net.Conn, msg string) {
	var outBuffer bytes.Buffer
	_, err := outBuffer.WriteString(msg)
	if err != nil {
		logging.Error("Failed writing string to buffer", err.Error())
		return
	}
	_, err = socket.Write(outBuffer.Bytes())
	if err != nil {
		logging.Error("Failed writing string to socket", err.Error())
		return
	}
}
