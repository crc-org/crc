package websocket

import (
	"sync"
)

type Endpoint interface {
	addClient(client *wsClient)
	deleteClient(client *wsClient)
	setHandler(handler *EndpointHandler)
	Write([]byte)
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
func (e *endpoint) Write(data []byte) {
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
}

func (e *endpoint) setHandler(handler *EndpointHandler) {
	e.handler = handler
}
