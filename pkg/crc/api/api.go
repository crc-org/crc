package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"

	"github.com/code-ready/crc/pkg/crc/logging"
)

func CreateServer(socketPath string, config newConfigFunc, machine newMachineFunc) (Server, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		logging.Error("Failed to create socket: ", err.Error())
		return Server{}, err
	}
	return createServerWithListener(listener, config, machine)
}

func createServerWithListener(listener net.Listener, config newConfigFunc, machine newMachineFunc) (Server, error) {
	apiServer := Server{
		listener:               listener,
		clusterOpsRequestsChan: make(chan clusterOpsRequest, 10),
		handlerFactory: func() (RequestHandler, error) {
			cfg, err := config()
			if err != nil {
				return nil, err
			}
			return &Handler{
				Config:        cfg,
				MachineClient: &Adapter{Underlying: machine(cfg)},
			}, nil
		},
	}
	return apiServer, nil
}

func (api Server) Serve() error {
	go api.handleClusterOperations() // go routine that handles start, stop and delete calls
	for {
		conn, err := api.listener.Accept()
		if err != nil {
			logging.Error("Error establishing communication: ", err.Error())
			continue
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

	handler, err := api.handlerFactory()
	if err != nil {
		logging.Error(err.Error())
		result = encodeErrorToJSON(fmt.Sprintf("Failed to initialize new config store: %v", err))
		writeStringToSocket(conn, result)
		return
	}

	switch req.Command {
	case "start":
		result = handler.Start(req.Args)
	case "stop":
		result = handler.Stop()
	case "status":
		result = handler.Status()
	case "delete":
		result = handler.Delete()
	case "version":
		result = handler.GetVersion()
	case "setconfig":
		result = handler.SetConfig(req.Args)
	case "unsetconfig":
		result = handler.UnsetConfig(req.Args)
	case "getconfig":
		result = handler.GetConfig(req.Args)
	case "webconsoleurl":
		result = handler.GetWebconsoleInfo()
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
	logging.Debug("Received Request:", string(inBuffer[0:numBytes]))
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

	case "status", "version", "setconfig", "getconfig", "unsetconfig", "webconsoleurl":
		go api.handleRequest(req, conn)

	default:
		err := encodeErrorToJSON(fmt.Sprintf("Unknown command supplied: %s", req.Command))
		writeStringToSocket(conn, err)
		conn.Close()
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

func addRequestToChannel(req clusterOpsRequest, requestsChan chan clusterOpsRequest) bool {
	select {
	case requestsChan <- req:
		return true
	default:
		return false
	}
}
