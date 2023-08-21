package events

import (
	"net/http"
	"sync"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/r3labs/sse/v2"
)

type EventServer struct {
	sseServer *sse.Server
	muStreams sync.RWMutex
	streams   map[string]EventStream
	machine   machine.Client
}

func NewEventServer(machine machine.Client) *EventServer {

	var sseServer = sse.New()
	sseServer.AutoReplay = false

	eventServer := &EventServer{
		sseServer: sseServer,
		machine:   machine,
		streams:   map[string]EventStream{},
	}

	sseServer.OnSubscribe = func(streamId string, sub *sse.Subscriber) {
		logging.Debugf("OnSubscribe on channel: %s", streamId)
		stream, ok := eventServer.streams[streamId]
		eventServer.muStreams.Lock()
		defer eventServer.muStreams.Unlock()
		if !ok {
			stream = createEventStream(eventServer, streamId)
			if stream == nil {
				logging.Errorf("Could not create EventStream for %s", streamId)
				return
			}
			eventServer.streams[streamId] = stream
		}

		stream.AddSubscriber(sub)
	}

	sseServer.OnUnsubscribe = func(streamId string, sub *sse.Subscriber) {
		logging.Debugf("OnUnsubscribe on channel: %s", streamId)
		stream, ok := eventServer.streams[streamId]
		if !ok {
			logging.Debugf("Could not find stream:%s", streamId)
			return
		}
		stream.RemoveSubscriber(sub)
	}

	sseServer.CreateStream(LOGS)
	sseServer.CreateStream(STATUS)
	return eventServer
}

func (es *EventServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	es.sseServer.ServeHTTP(w, r)
}

func createEventStream(server *EventServer, streamID string) EventStream {
	switch streamID {
	case LOGS:
		return newLogsStream(server)
	case STATUS:
		return newStatusStream(server)
	}
	return nil
}
