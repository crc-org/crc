package events

import (
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
)

type StatusChangedEvent struct {
	State state.State
	Error error
}

var (
	StatusChanged = NewEvent[StatusChangedEvent]()
)
