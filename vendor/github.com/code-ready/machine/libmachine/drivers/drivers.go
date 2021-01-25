package drivers

import (
	"errors"

	"github.com/code-ready/machine/libmachine/state"
	log "github.com/sirupsen/logrus"
)

// Driver defines how a host is created and controlled. Different types of
// driver represent different ways hosts can be created (e.g. different
// hypervisors, different cloud providers)
type Driver interface {
	// Create a host using the driver's config
	Create() error

	// DriverName returns the name of the driver
	DriverName() string

	// GetIP returns an IP or hostname that this host is available at
	// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
	GetIP() (string, error)

	// GetMachineName returns the name of the machine
	GetMachineName() string

	// GetBundleName() Returns the name of the unpacked bundle which was used to create this machine
	GetBundleName() (string, error)

	// GetState returns the state that the host is in (running, stopped, etc)
	GetState() (state.State, error)

	// Kill stops a host forcefully
	Kill() error

	// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
	PreCreateCheck() error

	// Remove a host
	Remove() error

	// UpdateConfigRaw allows to change the state (memory, ...) of an already created machine
	UpdateConfigRaw(rawDriver []byte) error

	// Start a host
	Start() error

	// Stop a host gracefully
	Stop() error

	// Get Version information
	DriverVersion() string
}

var ErrHostIsNotRunning = errors.New("Host is not running")
var ErrNotImplemented = errors.New("Not Implemented")

func MachineInState(d Driver, desiredState state.State) func() bool {
	return func() bool {
		currentState, err := d.GetState()
		if err != nil {
			log.Debugf("Error getting machine state: %s", err)
		}
		if currentState == desiredState {
			return true
		}
		return false
	}
}
