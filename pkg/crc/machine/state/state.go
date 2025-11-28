package state

// State represents the state of crc (both VM and components)
type State string

const (
	Running  State = "Running"
	Stopped  State = "Stopped"
	Stopping State = "Stopping"
	Starting State = "Starting"
	Error    State = "Error"
)
