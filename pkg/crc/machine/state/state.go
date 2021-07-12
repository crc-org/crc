package state

import libmachinestate "github.com/code-ready/machine/libmachine/state"

// State represents the state of crc (both VM and components)
type State string

const (
	Running  State = "Running"
	Stopped  State = "Stopped"
	Stopping State = "Stopping"
	Starting State = "Starting"
	Error    State = "Error"
)

func FromMachine(input libmachinestate.State) State {
	switch input {
	case libmachinestate.Running:
		return Running
	case libmachinestate.Stopped:
		return Stopped
	}
	return Error
}
