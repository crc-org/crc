package events

import "github.com/r3labs/sse/v2"

const (
	LOGS   = "logs"   // Logs event channel, contains daemon logs
	STATUS = "status" // status event channel, contains VM load info
)

type EventPublisher interface {
	Publish(event *sse.Event)
}

type EventProducer interface {
	Start(publisher EventPublisher)
	Stop()
}

type eventPublisher struct {
	streamID  string
	sseServer *sse.Server
}

func newEventPublisher(streamID string, server *sse.Server) EventPublisher {
	return &eventPublisher{
		streamID:  streamID,
		sseServer: server,
	}
}

func (ep *eventPublisher) Publish(event *sse.Event) {
	ep.sseServer.Publish(ep.streamID, event)
}
