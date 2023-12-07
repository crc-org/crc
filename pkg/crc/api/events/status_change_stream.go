package events

import (
	"encoding/json"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	"github.com/crc-org/crc/v2/pkg/events"
	"github.com/r3labs/sse/v2"
)

type statusChangeEvent struct {
	Status *types.ClusterStatusResult `json:"status"`
	Error  string                     `json:"error,omitempty"`
}

type statusChangeListener struct {
	machineClient machine.Client
	publisher     EventPublisher
}

func newStatusChangeStream(server *EventServer) EventStream {
	return newStream(newStatusChangeListener(server.machine), newEventPublisher(StatusChange, server.sseServer))
}

func newStatusChangeListener(client machine.Client) EventProducer {
	return &statusChangeListener{
		machineClient: client,
	}
}

func (st *statusChangeListener) Notify(changedEvent events.StatusChangedEvent) {
	logging.Debugf("State Changed Event %s", changedEvent)
	var event statusChangeEvent
	status, err := st.machineClient.Status()
	// if we cannot receive actual state, send error state with error description
	if err != nil {
		event = statusChangeEvent{Status: &types.ClusterStatusResult{
			CrcStatus: state.Error,
		}, Error: err.Error()}
	} else {
		// event could be fired, before actual code, which change state is called
		// so status could contain 'old' state, replace it with state received in event
		status.CrcStatus = changedEvent.State // override with actual reported state
		event = statusChangeEvent{Status: status}
		if changedEvent.Error != nil {
			event.Error = changedEvent.Error.Error()
		}

	}
	data, err := json.Marshal(event)
	if err != nil {
		logging.Errorf("Could not serealize status changed event in to JSON: %s", err)
		return
	}
	st.publisher.Publish(&sse.Event{Event: []byte(StatusChange), Data: data})
}

func (st *statusChangeListener) Start(publisher EventPublisher) {
	st.publisher = publisher
	events.StatusChanged.AddListener(st)

}

func (st *statusChangeListener) Stop() {
	events.StatusChanged.RemoveListener(st)
}
