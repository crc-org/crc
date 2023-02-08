package websocket

import (
	"encoding/json"
	"io"
	"time"

	"github.com/crc-org/crc/pkg/crc/logging"
	crcMachine "github.com/crc-org/crc/pkg/crc/machine"
)

type genData func() (interface{}, error)

// A data generator for a websocket endpoint. It will fetch data at regular intervals, and
// send it to all clients connected to the endpoint.
type TickListener struct {
	done       chan bool
	generator  genData
	tickPeriod time.Duration
}

func NewStatusListener(machine crcMachine.Client) ConnectionListener {
	getStatus := func() (interface{}, error) {
		return machine.GetClusterLoad()
	}
	return NewTickListener(getStatus)
}

func NewTickListener(generator genData) ConnectionListener {
	return &TickListener{
		done:       make(chan bool),
		generator:  generator,
		tickPeriod: 2000 * time.Millisecond,
	}
}

func (s *TickListener) start(dataSender io.Writer) {

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
				_, err = dataSender.Write(bytes)
				if err != nil {
					logging.Errorf("unexpected error during writing data to WebSocket: %v", err)
				}
			}
		}
	}()

}

func (s *TickListener) stop() {
	s.done <- true
}
