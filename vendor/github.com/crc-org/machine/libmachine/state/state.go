package state

// State represents the state of a host
type State int

const (
	reserved0 State = iota //nolint:deadcode,varcheck
	Running
	reserved2 //nolint:deadcode,varcheck
	reserved3 //nolint:deadcode,varcheck
	Stopped
	reserved5 //nolint:deadcode,varcheck
	reserved6 //nolint:deadcode,varcheck
	Error
	reserved8 //nolint:deadcode,varcheck
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
