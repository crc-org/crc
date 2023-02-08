package websocket

import (
	"io"
)

// An EndpointHandler tells a set of listeners when they should be generating
// data for sending to clients.
type EndpointHandler struct {
	listeners  []ConnectionListener
	dataSender io.Writer
}

type ConnectionListener interface {
	// called when first client connected
	start(dataSender io.Writer)
	// called when all clients close connections
	stop()
}

func NewEndpointHandler(dataSender io.Writer) *EndpointHandler {
	handler := &EndpointHandler{
		listeners:  make([]ConnectionListener, 0),
		dataSender: dataSender,
	}

	return handler
}

func (h *EndpointHandler) hasClient() {
	for _, l := range h.listeners {
		l.start(h.dataSender)
	}
}

func (h *EndpointHandler) noClient() {
	for _, l := range h.listeners {
		l.stop()
	}
}

func (h *EndpointHandler) addListener(listener ConnectionListener) {
	h.listeners = append(h.listeners, listener)
}
