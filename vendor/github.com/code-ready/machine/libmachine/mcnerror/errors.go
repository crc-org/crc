package mcnerror

import (
	"errors"
	"fmt"
	"strings"

	"github.com/code-ready/machine/libmachine/state"
)

var (
	ErrInvalidHostname = errors.New("Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -")
)

type ErrHostDoesNotExist struct {
	Name string
}

func (e ErrHostDoesNotExist) Error() string {
	return fmt.Sprintf("Machine %q does not exist. Use \"crc start\" to add a new one.", e.Name)
}

type ErrHostAlreadyExists struct {
	Name string
}

func (e ErrHostAlreadyExists) Error() string {
	return fmt.Sprintf("Machine %q already exists", e.Name)
}

type ErrDuringPreCreate struct {
	Cause error
}

func (e ErrDuringPreCreate) Error() string {
	return fmt.Sprintf("Error with pre-create check: %q", e.Cause)
}

type ErrHostAlreadyInState struct {
	Name  string
	State state.State
}

func (e ErrHostAlreadyInState) Error() string {
	return fmt.Sprintf("Machine %q is already %s.", e.Name, strings.ToLower(e.State.String()))
}
