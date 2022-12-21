package websocket

import (
	"encoding/json"
	"time"

	"github.com/crc-org/crc/pkg/crc/logging"
	"github.com/crc-org/crc/pkg/crc/machine"
)

type StatusConnectionListener struct {
	machine machine.Client
	done    chan bool
}

func NewStatusListener(machine machine.Client) ConnectionListener {
	return &StatusConnectionListener{
		machine: machine,
		done:    make(chan bool),
	}
}

func (s StatusConnectionListener) start(sendData SendData) {

	ticker := time.NewTicker(2000 * time.Millisecond)
	go func() {
		for {
			select {
			case <-s.done:
				ticker.Stop()
				logging.Debug("stop fetching machine info")
				return
			case <-ticker.C:
				status, err := s.machine.GetClusterLoad()
				if err != nil {
					logging.Errorf("unexpected error during getting machine status: %v", err)
				}

				bytes, marshallError := json.Marshal(status)
				if marshallError != nil {
					logging.Errorf("unexpected error during status object to JSON conversion: %v", err)
					continue
				}
				sendData(bytes)
			}
		}
	}()

}

func (s StatusConnectionListener) stop() {
	s.done <- true
}
