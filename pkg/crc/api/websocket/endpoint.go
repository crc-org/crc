package websocket

import (
	"io"
	"sync"
)

// websocket endpoint such as "/status". It has a number of clients connected
// to it, and works together with an EndpointHandler to listen for the data to
// send to its clients.
type Endpoint interface {
	addClient(client *wsClient)
	deleteClient(client *wsClient)
	setHandler(handler *EndpointHandler)
	io.Writer
}

type endpoint struct {
	clients      map[*wsClient]struct{}
	clientsMutex sync.Mutex
	handler      *EndpointHandler
}

func NewEndpoint() Endpoint {
	return &endpoint{
		clients: make(map[*wsClient]struct{}),
	}
}

func (e *endpoint) addClient(client *wsClient) {
	e.clientsMutex.Lock()
	defer e.clientsMutex.Unlock()

	e.clients[client] = struct{}{}
	if len(e.clients) == 1 {
		e.handler.hasClient()
	}

}

func (e *endpoint) deleteClient(client *wsClient) {
	e.clientsMutex.Lock()
	defer e.clientsMutex.Unlock()

	delete(e.clients, client)
	// all clients disconnected
	if len(e.clients) == 0 {
		e.handler.noClient()
	}
}

// send data bytes to clients
func (e *endpoint) Write(data []byte) (int, error) {
	e.clientsMutex.Lock()
	defer e.clientsMutex.Unlock()

	// use 'msgs' client channel to pass data to client impl
	for client := range e.clients {
		select {
		case client.msgs <- data:
		default:
			go client.closeSlow()
		}
	}

	return len(data), nil
}

func (e *endpoint) setHandler(handler *EndpointHandler) {
	e.handler = handler
}
