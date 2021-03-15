package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	crcConfig "github.com/code-ready/crc/pkg/crc/config"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine"
)

func CreateServer(socketPath string, config crcConfig.Storage, machine machine.Client, logger Logger) (Server, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logging.Error("Failed to create socket: ", err.Error())
		return Server{}, err
	}
	return createServerWithListener(listener, config, machine, logger)
}

func createServerWithListener(listener net.Listener, config crcConfig.Storage, machine machine.Client, logger Logger) (Server, error) {
	apiServer := Server{
		listener:               listener,
		clusterOpsRequestsChan: make(chan clusterOpsRequest, 10),
		handler: &Handler{
			Config:        config,
			MachineClient: &Adapter{Underlying: machine},
			Logger:        logger,
		},
	}
	return apiServer, nil
}

func (api Server) Serve() error {
	go api.handleClusterOperations() // go routine that handles start, stop and delete calls
	for {
		conn, err := api.listener.Accept()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				logging.Errorf("accept temporary error: %v", err)
				continue
			}
			return err
		}
		api.handleConnections(conn) // handle version, status, webconsole, etc. requests
	}
}

func (api Server) handleClusterOperations() {
	for req := range api.clusterOpsRequestsChan {
		api.handleRequest(req.command, req.socket)
	}
}

func (api Server) handleRequest(req commandRequest, conn net.Conn) {
	defer conn.Close()
	var result string

	switch req.Command {
	case "start":
		result = api.handler.Start(req.Args)
	case "stop":
		result = api.handler.Stop()
	case "status":
		result = api.handler.Status()
	case "delete":
		result = api.handler.Delete()
	case "version":
		result = api.handler.GetVersion()
	case "setconfig":
		result = api.handler.SetConfig(req.Args)
	case "unsetconfig":
		result = api.handler.UnsetConfig(req.Args)
	case "getconfig":
		result = api.handler.GetConfig(req.Args)
	case "webconsoleurl":
		result = api.handler.GetWebconsoleInfo()
	case "logs":
		result = api.handler.Logs()
	default:
		result = encodeErrorToJSON(fmt.Sprintf("Unknown command supplied: %s", req.Command))
	}
	writeStringToSocket(conn, result)
}

func (api Server) handleConnections(conn net.Conn) {
	inBuffer := make([]byte, 1024)
	var req commandRequest
	numBytes, err := conn.Read(inBuffer)
	if err != nil || numBytes == 0 || numBytes == cap(inBuffer) {
		logging.Error("Error reading from socket")
		return
	}
	logging.Debugf("rpc:%p: Received request: %s", conn, string(inBuffer[0:numBytes]))
	err = json.Unmarshal(inBuffer[0:numBytes], &req)
	if err != nil {
		logging.Error("Error decoding request: ", err.Error())
		return
	}
	// start, stop and delete are slow operations, and change the VM state so they have to run sequentially.
	// We don't want other operations querying the status of the VM to be blocked by these,
	// so they are treated by a dedicated go routine

	switch req.Command {
	case "start", "stop", "delete":
		// queue new request to channel
		r := clusterOpsRequest{
			command: req,
			socket:  conn,
		}
		if !addRequestToChannel(r, api.clusterOpsRequestsChan) {
			logging.Error("Channel capacity reached, unable to add new request")
			errMsg := encodeErrorToJSON("Sockets channel capacity reached, unable to add new request")
			writeStringToSocket(conn, errMsg)
			conn.Close()
		}

	case "status", "version", "setconfig", "getconfig", "unsetconfig", "webconsoleurl", "logs":
		go api.handleRequest(req, conn)

	default:
		err := encodeErrorToJSON(fmt.Sprintf("Unknown command supplied: %s", req.Command))
		writeStringToSocket(conn, err)
		conn.Close()
	}
}

func writeStringToSocket(socket net.Conn, msg string) {
	logging.Debugf("rpc:%p: Sending answer: %s", socket, msg)

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

func addRequestToChannel(req clusterOpsRequest, requestsChan chan clusterOpsRequest) bool {
	select {
	case requestsChan <- req:
		return true
	default:
		return false
	}
}
