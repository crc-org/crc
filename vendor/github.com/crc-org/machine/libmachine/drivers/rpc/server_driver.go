package rpcdriver

import (
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/crc-org/machine/libmachine/drivers"
	"github.com/crc-org/machine/libmachine/state"
	"github.com/crc-org/machine/libmachine/version"
	log "github.com/sirupsen/logrus"
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

func (r *RPCServerDriver) logRPC(op string, level log.Level, extra log.Fields) {
	fields := log.Fields{"operation": op}
	if r.ActualDriver != nil {
		func() {
			defer func() { _ = recover() }()
			if name := r.ActualDriver.GetMachineName(); name != "" {
				fields["machine"] = name
			}
		}()
	}
	for k, v := range extra {
		fields[k] = v
	}
	log.WithFields(fields).Log(level, "RPC server invocation")
}

func (r *RPCServerDriver) Close(_, _ *struct{}) error {
	r.logRPC("Close", log.InfoLevel, nil)
	r.CloseCh <- true
	return nil
}

func (r *RPCServerDriver) GetVersion(_ *struct{}, reply *int) error {
	r.logRPC("GetVersion", log.InfoLevel, nil)
	*reply = version.APIVersion
	return nil
}

func (r *RPCServerDriver) GetConfigRaw(_ *struct{}, reply *[]byte) error {
	driverData, err := json.Marshal(r.ActualDriver)
	if err != nil {
		return err
	}

	*reply = driverData
	r.logRPC("GetConfigRaw", log.InfoLevel, log.Fields{"config_bytes": len(driverData)})

	return nil
}

func (r *RPCServerDriver) UpdateConfigRaw(data []byte, _ *struct{}) error {
	r.logRPC("UpdateConfigRaw", log.InfoLevel, log.Fields{"config_bytes": len(data)})
	return r.ActualDriver.UpdateConfigRaw(data)
}

func (r *RPCServerDriver) SetConfigRaw(data []byte, _ *struct{}) error {
	r.logRPC("SetConfigRaw", log.InfoLevel, log.Fields{"config_bytes": len(data)})
	return json.Unmarshal(data, &r.ActualDriver)
}

func trapPanic(err *error) {
	if r := recover(); r != nil {
		*err = fmt.Errorf("Panic in the driver: %s\n%s", r.(error), stdStacker.Stack())
	}
}

func (r *RPCServerDriver) Create(_, _ *struct{}) (err error) {
	r.logRPC("Create", log.InfoLevel, nil)
	// In an ideal world, plugins wouldn't ever panic.  However, panics
	// have been known to happen and cause issues.  Therefore, we recover
	// and do not crash the RPC server completely in the case of a panic
	// during create.
	defer trapPanic(&err)

	err = r.ActualDriver.Create()

	return err
}

func (r *RPCServerDriver) DriverName(_ *struct{}, reply *string) error {
	r.logRPC("DriverName", log.InfoLevel, nil)
	*reply = r.ActualDriver.DriverName()
	return nil
}

func (r *RPCServerDriver) GetIP(_ *struct{}, reply *string) error {
	r.logRPC("GetIP", log.InfoLevel, nil)
	ip, err := r.ActualDriver.GetIP()
	*reply = ip
	return err
}

func (r *RPCServerDriver) GetMachineName(_ *struct{}, reply *string) error {
	r.logRPC("GetMachineName", log.InfoLevel, nil)
	*reply = r.ActualDriver.GetMachineName()
	return nil
}

func (r *RPCServerDriver) GetBundleName(_ *struct{}, reply *string) error {
	r.logRPC("GetBundleName", log.InfoLevel, nil)
	path, err := r.ActualDriver.GetBundleName()
	*reply = path
	return err
}

func (r *RPCServerDriver) GetState(_ *struct{}, reply *state.State) error {
	r.logRPC("GetState", log.InfoLevel, nil)
	s, err := r.ActualDriver.GetState()
	*reply = s
	return err
}

func (r *RPCServerDriver) Kill(_ *struct{}, _ *struct{}) error {
	r.logRPC("Kill", log.InfoLevel, nil)
	return r.ActualDriver.Kill()
}

func (r *RPCServerDriver) PreCreateCheck(_ *struct{}, _ *struct{}) error {
	r.logRPC("PreCreateCheck", log.InfoLevel, nil)
	return r.ActualDriver.PreCreateCheck()
}

func (r *RPCServerDriver) Remove(_ *struct{}, _ *struct{}) error {
	r.logRPC("Remove", log.InfoLevel, nil)
	return r.ActualDriver.Remove()
}

func (r *RPCServerDriver) Start(_ *struct{}, _ *struct{}) error {
	r.logRPC("Start", log.InfoLevel, nil)
	return r.ActualDriver.Start()
}

func (r *RPCServerDriver) Stop(_ *struct{}, _ *struct{}) error {
	r.logRPC("Stop", log.InfoLevel, nil)
	return r.ActualDriver.Stop()
}

func (r *RPCServerDriver) Heartbeat(_ *struct{}, _ *struct{}) error {
	r.HeartbeatCh <- true
	return nil
}

func (r *RPCServerDriver) GetSharedDirs(_ *struct{}, reply *[]drivers.SharedDir) error {
	sharedDirs, err := r.ActualDriver.GetSharedDirs()
	*reply = sharedDirs
	if err == nil {
		r.logRPC("GetSharedDirs", log.InfoLevel, log.Fields{"shared_dirs": len(sharedDirs)})
	}
	return err
}
