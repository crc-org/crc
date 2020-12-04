package rpcdriver

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/code-ready/machine/libmachine/version"
)

type Stacker interface {
	Stack() []byte
}

type StandardStack struct{}

func (ss *StandardStack) Stack() []byte {
	return debug.Stack()
}

var (
	stdStacker Stacker = &StandardStack{}
)

type RPCServerDriver struct {
	ActualDriver drivers.Driver
	CloseCh      chan bool
	HeartbeatCh  chan bool
}

func NewRPCServerDriver(d drivers.Driver) *RPCServerDriver {
	return &RPCServerDriver{
		ActualDriver: d,
		CloseCh:      make(chan bool),
		HeartbeatCh:  make(chan bool),
	}
}

func (r *RPCServerDriver) Close(_, _ *struct{}) error {
	r.CloseCh <- true
	return nil
}

func (r *RPCServerDriver) GetVersion(_ *struct{}, reply *int) error {
	*reply = version.APIVersion
	return nil
}

func (r *RPCServerDriver) GetConfigRaw(_ *struct{}, reply *[]byte) error {
	driverData, err := json.Marshal(r.ActualDriver)
	if err != nil {
		return err
	}

	*reply = driverData

	return nil
}

func (r *RPCServerDriver) UpdateConfigRaw(data []byte, _ *struct{}) error {
	return r.ActualDriver.UpdateConfigRaw(data)
}

func (r *RPCServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	return json.Unmarshal(data, &r.ActualDriver)
}

func trapPanic(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("Panic in the driver: %s\n%s", r.(error), stdStacker.Stack())
	}
}

func (r *RPCServerDriver) Create(_, _ *struct{}) (err error) {
	// In an ideal world, plugins wouldn't ever panic.  However, panics
	// have been known to happen and cause issues.  Therefore, we recover
	// and do not crash the RPC server completely in the case of a panic
	// during create.
	defer trapPanic(&err)

	err = r.ActualDriver.Create()

	return err
}

func (r *RPCServerDriver) DriverName(_ *struct{}, reply *string) error {
	*reply = r.ActualDriver.DriverName()
	return nil
}

func (r *RPCServerDriver) GetIP(_ *struct{}, reply *string) error {
	ip, err := r.ActualDriver.GetIP()
	*reply = ip
	return err
}

func (r *RPCServerDriver) GetMachineName(_ *struct{}, reply *string) error {
	*reply = r.ActualDriver.GetMachineName()
	return nil
}

func (r *RPCServerDriver) GetBundleName(_ *struct{}, reply *string) error {
	path, err := r.ActualDriver.GetBundleName()
	*reply = path
	return err
}

func (r *RPCServerDriver) GetState(_ *struct{}, reply *state.State) error {
	s, err := r.ActualDriver.GetState()
	*reply = s
	return err
}

func (r *RPCServerDriver) Kill(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Kill()
}

func (r *RPCServerDriver) PreCreateCheck(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.PreCreateCheck()
}

func (r *RPCServerDriver) Remove(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Remove()
}

func (r *RPCServerDriver) Start(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Start()
}

func (r *RPCServerDriver) Stop(_ *struct{}, _ *struct{}) error {
	return r.ActualDriver.Stop()
}

func (r *RPCServerDriver) Heartbeat(_ *struct{}, _ *struct{}) error {
	r.HeartbeatCh <- true
	return nil
}
