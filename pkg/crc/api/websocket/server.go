package websocket

import (
	goContext "context"
	"errors"
	"net/http"
	"time"

	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/crc-org/crc/pkg/crc/machine"
	"nhooyr.io/websocket"
)

type wsClient struct {
	msgs      chan []byte
	closeSlow func()
}

type Server struct {
	// clientMessageBuffer controls the max number
	// of messages that can be queued for a subscriber
	// before it is kicked.
	//
	// Defaults to 16.
	clientMessageBuffer int

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	endpoints map[string]Endpoint
}

func NewWebsocketServer(machine machine.Client) *Server {
	ws := &Server{
		clientMessageBuffer: 16,
		endpoints:           map[string]Endpoint{},
	}

	ws.endpoints["/status"] = createStatusEndpoint(machine)

	ws.serveMux.HandleFunc("/status", func(writer http.ResponseWriter, request *http.Request) {
		ws.handleEndpoint("/status", writer, request)
	})

	return ws
}

func (ws *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.serveMux.ServeHTTP(w, r)
}

func (ws *Server) handleEndpoint(endpointPath string, w http.ResponseWriter, r *http.Request) {
	logging.Debugf("handle %s request", endpointPath)
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		logging.Infof("%v", err)
		return
	}

	defer c.Close(websocket.StatusInternalError, "")

	// TODO add reading messages from client

	err = ws.subscribe(endpointPath, r.Context(), c)
	if errors.Is(err, goContext.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		logging.Infof("%v", err)
		return
	}
}

func (ws *Server) subscribe(endpointPath string, ctx goContext.Context, c *websocket.Conn) error {
	logging.Debugf("Subscribe: %s", endpointPath)
	ctx = c.CloseRead(ctx)
	client := &wsClient{
		msgs: make(chan []byte, ws.clientMessageBuffer),
		closeSlow: func() {
			c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}

	ws.addClient(endpointPath, client)
	defer ws.deleteClient(endpointPath, client)

	// pull data from 'msgs' client channel and write to socket
	for {
		select {
		case msg := <-client.msgs:
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// addClient registers a client.
func (ws *Server) addClient(endpointPath string, s *wsClient) {
	ws.endpoints[endpointPath].addClient(s)
}

// deleteClient deletes the given client.
func (ws *Server) deleteClient(endpointPath string, s *wsClient) {
	ws.endpoints[endpointPath].deleteClient(s)
}

func writeTimeout(ctx goContext.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := goContext.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

func createStatusEndpoint(machine machine.Client) Endpoint {
	statusEndpoint := NewEndpoint()
	statusEndpointHandler := NewEndpointHandler(statusEndpoint.Write)
	statusEndpoint.setHandler(statusEndpointHandler)
	statusEndpointHandler.addListener(NewStatusListener(machine))
	return statusEndpoint
}
