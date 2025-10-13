package machine

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/crc-org/crc/v2/pkg/crc/logging"
	"github.com/crc-org/crc/v2/pkg/crc/machine/state"
	"github.com/crc-org/crc/v2/pkg/crc/machine/types"
	crcPreset "github.com/crc-org/crc/v2/pkg/crc/preset"
	"github.com/crc-org/crc/v2/pkg/events"
)

const startCancelTimeout = 15 * time.Second

type State string

const (
	Idle     State = "Idle"
	Deleting State = "Deleting"
	Stopping State = "Stopping"
	Starting State = "Starting"
)

type Synchronized struct {
	underlying Client

	stateLock    sync.Mutex
	currentState State
	startCancel  context.CancelFunc

	syncOperationDone chan State
}

func NewSynchronizedMachine(machine Client) *Synchronized {
	return &Synchronized{
		underlying:        machine,
		currentState:      Idle,
		syncOperationDone: make(chan State, 1),
	}
}

func (s *Synchronized) CurrentState() State {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	return s.currentStateUnlocked()
}

func (s *Synchronized) currentStateUnlocked() State {
	select {
	case st := <-s.syncOperationDone:
		if s.currentState == st {
			s.currentState = Idle
		}
		if st == Starting {
			s.startCancel = nil
		}
	default:
	}
	return s.currentState
}

func (s *Synchronized) Delete() error {
	if err := s.prepareStopDelete(Deleting); err != nil {
		return err
	}

	err := s.underlying.Delete()
	s.syncOperationDone <- Deleting

	if err == nil {
		events.StatusChanged.Fire(events.StatusChangedEvent{State: state.NoVM})
	}
	return err
}

func (s *Synchronized) prepareStart(startCancel context.CancelFunc) error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	if s.currentStateUnlocked() != Idle {
		return errors.New("cluster is busy")
	}
	s.startCancel = startCancel
	s.currentState = Starting
	events.StatusChanged.Fire(events.StatusChangedEvent{State: state.Starting})

	return nil
}

func (s *Synchronized) Start(ctx context.Context, startConfig types.StartConfig) (*types.StartResult, error) {
	ctx, startCancel := context.WithCancel(ctx)
	if err := s.prepareStart(startCancel); err != nil {
		return nil, err
	}

	startResult, err := s.underlying.Start(ctx, startConfig)
	s.syncOperationDone <- Starting

	if err == nil {
		events.StatusChanged.Fire(events.StatusChangedEvent{State: startResult.Status})
	} else {
		events.StatusChanged.Fire(events.StatusChangedEvent{State: state.Error, Error: err})
	}

	return startResult, err
}

/* cancel ongoing start, and wait until the start is fully cancelled. Time out if cancellation takes more than 'timeout'
 * s.stateLock must be locked before calling this function
 */
func (s *Synchronized) cancelUnlocked(timeout time.Duration) error {
	if s.startCancel != nil {
		logging.Infof("Cancelling virtual machine start...")
		s.startCancel()
	}
	select {
	case <-s.syncOperationDone:
		return nil
	case <-time.After(timeout):
		return errors.New("cannot abort startup sequence quickly enough")
	}
}

func (s *Synchronized) prepareStopDelete(state State) error {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	switch s.currentStateUnlocked() {
	case Starting:
		if err := s.cancelUnlocked(startCancelTimeout); err != nil {
			return err
		}
	case Idle:
		break
	case Deleting, Stopping:
		return errors.New("cluster is stopping or deleting")
	default:
		return errors.New("invalid condition")
	}

	s.currentState = state
	return nil
}

func (s *Synchronized) Stop() (state.State, error) {
	if err := s.prepareStopDelete(Stopping); err != nil {
		return state.Error, err
	}
	events.StatusChanged.Fire(events.StatusChangedEvent{State: state.Stopping})

	st, err := s.underlying.Stop()
	s.syncOperationDone <- Stopping

	if err == nil {
		events.StatusChanged.Fire(events.StatusChangedEvent{State: st})
	} else {
		events.StatusChanged.Fire(events.StatusChangedEvent{State: state.Error, Error: err})
	}
	return st, err
}

func (s *Synchronized) GetName() string {
	return s.underlying.GetName()
}

func (s *Synchronized) Exists() (bool, error) {
	return s.underlying.Exists()
}

func (s *Synchronized) GetConsoleURL() (*types.ConsoleResult, error) {
	return s.underlying.GetConsoleURL()
}

func (s *Synchronized) ConnectionDetails() (*types.ConnectionDetails, error) {
	return s.underlying.ConnectionDetails()
}

func (s *Synchronized) PowerOff() error {
	err := s.underlying.PowerOff()
	if err != nil {
		events.StatusChanged.Fire(events.StatusChangedEvent{State: state.Stopped})
	} else {
		events.StatusChanged.Fire(events.StatusChangedEvent{State: state.Error, Error: err})
	}

	return err
}

func (s *Synchronized) Status() (*types.ClusterStatusResult, error) {
	switch s.CurrentState() {
	case Starting:
		return &types.ClusterStatusResult{
			CrcStatus:       state.Starting,
			OpenshiftStatus: types.OpenshiftStarting,
			Preset:          s.underlying.GetPreset(),
		}, nil
	case Stopping, Deleting:
		return &types.ClusterStatusResult{
			CrcStatus:       state.Stopping,
			OpenshiftStatus: types.OpenshiftStopping,
			Preset:          s.underlying.GetPreset(),
		}, nil
	default:
		return s.underlying.Status()
	}
}

func (s *Synchronized) GetClusterLoad() (*types.ClusterLoadResult, error) {
	return s.underlying.GetClusterLoad()
}

func (s *Synchronized) IsRunning() (bool, error) {
	return s.underlying.IsRunning()
}

func (s *Synchronized) GenerateBundle(forceStop bool) error {
	return s.underlying.GenerateBundle(forceStop)
}

func (s *Synchronized) GetPreset() crcPreset.Preset {
	return s.underlying.GetPreset()
}
