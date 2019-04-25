package states

import (
	"strings"
)

type State int

const (
	Unknown State = iota
	Running
	Stopped
	NotFound
	Error
)

var states = []string{
	"unknown",
	"active (running)",
	"inactive (dead)",
	"could not be found",
	"error",
}

func (s State) String() string {
	if int(s) >= 0 && int(s) < len(states) {
		return states[s]
	}
	return ""
}

func Compare(input string) State {
	if strings.Contains(input, states[Running]) {
		return Running
	}
	if strings.Contains(input, states[Stopped]) {
		return Stopped
	}
	if strings.Contains(input, states[NotFound]) {
		return NotFound
	}
	return Unknown
}
