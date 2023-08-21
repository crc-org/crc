package events

import (
	"encoding/json"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	crcMachine "github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/r3labs/sse/v2"
)

type genData func() (interface{}, error)

// TickListener a data generator for an event stream. It will fetch data at regular intervals, and
// send it to all clients connected to the endpoint.
type TickListener struct {
	done       chan bool
	generator  genData
	tickPeriod time.Duration
}

func newStatusStream(server *EventServer) EventStream {
	return newStream(NewStatusListener(server.machine), newEventPublisher(STATUS, server.sseServer))
}

func NewStatusListener(machine crcMachine.Client) EventProducer {
	getStatus := func() (interface{}, error) {
		return machine.GetClusterLoad()
	}
	return NewTickListener(getStatus)
}

func NewTickListener(generator genData) EventProducer {
	return &TickListener{
		done:       make(chan bool),
		generator:  generator,
		tickPeriod: 2000 * time.Millisecond,
	}
}

func (s *TickListener) Start(publisher EventPublisher) {
	logging.Debug("Start sending status events")
	ticker := time.NewTicker(s.tickPeriod)
	go func() {
		for {
			select {
			case <-s.done:
				ticker.Stop()
				logging.Debug("stop fetching machine info")
				return
			case <-ticker.C:
				data, err := s.generator()
				if err != nil {
					logging.Errorf("unexpected error during getting machine status: %v", err)
					continue
				}

				bytes, marshallError := json.Marshal(data)
				if marshallError != nil {
					logging.Errorf("unexpected error during status object to JSON conversion: %v", err)
					continue
				}
				publisher.Publish(&sse.Event{Event: []byte("status"), Data: bytes})
			}
		}
	}()
}

func (s *TickListener) Stop() {
	logging.Debug("Stop sending status events")
	s.done <- true
}
