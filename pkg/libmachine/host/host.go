package host

import (
	"errors"
	"fmt"
	"net/rpc"
	"strings"
	"time"

	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	log "github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
)

// ConfigVersion dictates which version of the config.json format is
// used. It needs to be bumped if there is a breaking change, and
// therefore migration, introduced to the config file format.
const Version = 3

type Host struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	DriverPath    string
	Name          string
	RawDriver     []byte `json:"-"`
}

type Metadata struct {
	ConfigVersion int
}

func (h *Host) runActionForState(action func() error, desiredState state.State) error {
	if err := MachineInState(h.Driver, desiredState)(); err == nil {
		return fmt.Errorf("machine is already %s", strings.ToLower(desiredState.String()))
	}

	if err := action(); err != nil {
		return err
	}

	return crcerrors.RetryAfter(3*time.Minute, MachineInState(h.Driver, desiredState), 3*time.Second)
}

func (h *Host) Start() error {
	log.Infof("Starting %q...", h.Name)
	if err := h.runActionForState(h.Driver.Start, state.Running); err != nil {
		return err
	}

	log.Infof("Machine %q was started.", h.Name)

	return nil
}

func (h *Host) Stop() error {
	log.Infof("Stopping %q...", h.Name)
	if err := h.runActionForState(h.Driver.Stop, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was stopped.", h.Name)
	return nil
}

func (h *Host) Kill() error {
	log.Infof("Killing %q...", h.Name)
	if err := h.runActionForState(h.Driver.Kill, state.Stopped); err != nil {
		return err
	}

	log.Infof("Machine %q was killed.", h.Name)
	return nil
}

func (h *Host) UpdateConfig(rawConfig []byte) error {
	err := h.Driver.UpdateConfigRaw(rawConfig)
	if err != nil {
		var e rpc.ServerError
		if errors.As(err, &e) && err.Error() == "Not Implemented" {
			err = drivers.ErrNotImplemented
		}
		return err
	}
	h.RawDriver = rawConfig

	return nil
}

func MachineInState(d drivers.Driver, desiredState state.State) func() error {
	return func() error {
		currentState, err := d.GetState()
		if err != nil {
			return err
		}
		if currentState == desiredState {
			return nil
		}
		return crcerrors.RetriableError{
			Err: fmt.Errorf("expected machine state %s, got %s", desiredState.String(), currentState.String()),
		}
	}
}
