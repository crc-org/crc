package state

// State represents the state of a host
type State int

const (
	reserved0 State = iota
	Running
	reserved2
	reserved3
	Stopped
	reserved5
	reserved6
	Error
	reserved8
)

var states = []string{
	"reserved0",
	"Running",
	"reserved2",
	"reserved3",
	"Stopped",
	"reserved5",
	"reserved6",
	"Error",
	"reserved8",
}

// Given a State type, returns its string representation
func (s State) String() string {
	if int(s) >= 0 && int(s) < len(states) {
		return states[s]
	}
	return ""
}
