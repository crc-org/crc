package host

import (
	"regexp"

	"github.com/code-ready/machine/libmachine/auth"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/engine"
	"github.com/code-ready/machine/libmachine/log"
	"github.com/code-ready/machine/libmachine/mcnerror"
	"github.com/code-ready/machine/libmachine/mcnutils"
	"github.com/code-ready/machine/libmachine/ssh"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/code-ready/machine/libmachine/swarm"
)

var (
	validHostNamePattern                  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]*$`)
	stdSSHClientCreator  SSHClientCreator = &StandardSSHClientCreator{}
)

type SSHClientCreator interface {
	CreateSSHClient(d drivers.Driver) (ssh.Client, error)
}

type StandardSSHClientCreator struct {
	drivers.Driver
}

func SetSSHClientCreator(creator SSHClientCreator) {
	stdSSHClientCreator = creator
}

type Host struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *Options
	Name          string
	RawDriver     []byte `json:"-"`
}

type Options struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.Options
	SwarmOptions  *swarm.Options
	AuthOptions   *auth.Options
}

type Metadata struct {
	ConfigVersion int
	DriverName    string
	HostOptions   Options
}

func ValidateHostName(name string) bool {
	return validHostNamePattern.MatchString(name)
}

func (h *Host) RunSSHCommand(command string) (string, error) {
	return drivers.RunSSHCommandFromDriver(h.Driver, command)
}

func (h *Host) CreateSSHClient() (ssh.Client, error) {
	return stdSSHClientCreator.CreateSSHClient(h.Driver)
}

func (creator *StandardSSHClientCreator) CreateSSHClient(d drivers.Driver) (ssh.Client, error) {
	addr, err := d.GetSSHHostname()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return &ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{}
	if d.GetSSHKeyPath() != "" {
		auth.Keys = []string{d.GetSSHKeyPath()}
	}

	return ssh.NewClient(d.GetSSHUsername(), addr, port, auth)
}

func (h *Host) runActionForState(action func() error, desiredState state.State) error {
	if drivers.MachineInState(h.Driver, desiredState)() {
		return mcnerror.ErrHostAlreadyInState{
			Name:  h.Name,
			State: desiredState,
		}
	}

	if err := action(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(h.Driver, desiredState))
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

func (h *Host) Restart() error {
	log.Infof("Restarting %q...", h.Name)
	if drivers.MachineInState(h.Driver, state.Stopped)() {
		if err := h.Start(); err != nil {
			return err
		}
	} else if drivers.MachineInState(h.Driver, state.Running)() {
		if err := h.Driver.Restart(); err != nil {
			return err
		}
		if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
			return err
		}
	}

	return nil
}

func (h *Host) URL() (string, error) {
	return h.Driver.GetURL()
}
