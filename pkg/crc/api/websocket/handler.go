package websocket

type EndpointHandler struct {
	listeners []ConnectionListener
	sendData  SendData
}

type SendData func([]byte)

type ConnectionListener interface {
	// called when first client connected
	start(sendData SendData)
	// called when all clients close connections
	stop()
}

func NewEndpointHandler(data SendData) *EndpointHandler {
	handler := &EndpointHandler{
		listeners: make([]ConnectionListener, 0),
		sendData:  data,
	}

	return handler
}

func (h *EndpointHandler) hasClient() {
	for _, l := range h.listeners {
		l.start(h.sendData)
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
