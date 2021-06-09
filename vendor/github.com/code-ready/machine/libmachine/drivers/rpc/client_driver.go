package rpcdriver

import (
	"fmt"
	"io"
	"net/rpc"
	"sync"
	"time"

	"github.com/code-ready/machine/libmachine/drivers/plugin/localbinary"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/code-ready/machine/libmachine/version"
	log "github.com/sirupsen/logrus"
)

var (
	heartbeatInterval = 5 * time.Second
)

type RPCClientDriverFactory interface {
	NewRPCClientDriver(driverName string, driverPath string, rawDriver []byte) (*RPCClientDriver, error)
	io.Closer
}

type DefaultRPCClientDriverFactory struct {
	openedDrivers     []*RPCClientDriver
	openedDriversLock sync.Locker
}

func NewRPCClientDriverFactory() RPCClientDriverFactory {
	return &DefaultRPCClientDriverFactory{
		openedDrivers:     []*RPCClientDriver{},
		openedDriversLock: &sync.Mutex{},
	}
}

type RPCClientDriver struct {
	plugin          io.Closer
	heartbeatDoneCh chan bool
	Client          *InternalClient
}

type RPCCall struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
}

type InternalClient struct {
	MachineName    string
	RPCClient      *rpc.Client
	rpcServiceName string
}

const (
	RPCServiceNameV0 = `RpcServerDriver`
	RPCServiceNameV1 = `RPCServerDriver`

	HeartbeatMethod       = `.Heartbeat`
	GetVersionMethod      = `.GetVersion`
	CloseMethod           = `.Close`
	SetConfigRawMethod    = `.SetConfigRaw`
	UpdateConfigRawMethod = `.UpdateConfigRaw`
	GetConfigRawMethod    = `.GetConfigRaw`
	DriverNameMethod      = `.DriverName`
	GetMachineNameMethod  = `.GetMachineName`
	GetIPMethod           = `.GetIP`
	GetBundleNameMethod   = `.GetBundleName`
	GetStateMethod        = `.GetState`
	PreCreateCheckMethod  = `.PreCreateCheck`
	CreateMethod          = `.Create`
	RemoveMethod          = `.Remove`
	StartMethod           = `.Start`
	StopMethod            = `.Stop`
	KillMethod            = `.Kill`
)

func (ic *InternalClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if serviceMethod != HeartbeatMethod {
		log.Debugf("(%s) Calling %+v", ic.MachineName, serviceMethod)
	}
	return ic.RPCClient.Call(ic.rpcServiceName+serviceMethod, args, reply)
}

func (ic *InternalClient) switchToV0() {
	ic.rpcServiceName = RPCServiceNameV0
}

func NewInternalClient(rpcclient *rpc.Client) *InternalClient {
	return &InternalClient{
		RPCClient:      rpcclient,
		rpcServiceName: RPCServiceNameV1,
	}
}

func (f *DefaultRPCClientDriverFactory) Close() error {
	f.openedDriversLock.Lock()
	defer f.openedDriversLock.Unlock()

	for _, openedDriver := range f.openedDrivers {
		// No need to display an error.
		// There's nothing we can do and it doesn't add value to the user.
		_ = openedDriver.close()
	}
	f.openedDrivers = []*RPCClientDriver{}

	return nil
}

func (f *DefaultRPCClientDriverFactory) NewRPCClientDriver(driverName string, driverPath string, rawDriver []byte) (*RPCClientDriver, error) {
	p, err := localbinary.NewPlugin(driverName, driverPath)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := p.Serve(); err != nil {
			// TODO: Is this best approach?
			log.Warnf(err.Error())
			return
		}
	}()

	addr, err := p.Address()
	if err != nil {
		return nil, fmt.Errorf("Error attempting to get plugin server address for RPC: %s", err)
	}

	rpcclient, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &RPCClientDriver{
		Client:          NewInternalClient(rpcclient),
		heartbeatDoneCh: make(chan bool),
	}

	f.openedDriversLock.Lock()
	f.openedDrivers = append(f.openedDrivers, c)
	f.openedDriversLock.Unlock()

	var serverVersion int
	if err := c.Client.Call(GetVersionMethod, struct{}{}, &serverVersion); err != nil {
		// this is the first call we make to the server. We try to play nice with old pre 0.5.1 client,
		// by gracefully trying old RPCServiceName, we do this only once, and keep the result for future calls.
		log.Debugf(err.Error())
		log.Debugf("Client (%s) with %s does not work, re-attempting with %s", c.Client.MachineName, RPCServiceNameV1, RPCServiceNameV0)
		c.Client.switchToV0()
		if err := c.Client.Call(GetVersionMethod, struct{}{}, &serverVersion); err != nil {
			return nil, err
		}
	}

	if serverVersion != version.APIVersion {
		return nil, fmt.Errorf("Driver binary uses an incompatible API version (%d)", serverVersion)
	}
	log.Debugf("Using API Version %d", serverVersion)

	go func(c *RPCClientDriver) {
		for {
			select {
			case <-c.heartbeatDoneCh:
				return
			case <-time.After(heartbeatInterval):
				if err := c.Client.Call(HeartbeatMethod, struct{}{}, nil); err != nil {
					log.Warnf("driver heartbeat failure: %v", err)
					if err := c.close(); err != nil {
						log.Warnf(err.Error())
					}
				}
			}
		}
	}(c)

	if err := c.SetConfigRaw(rawDriver); err != nil {
		return nil, err
	}

	mcnName := c.GetMachineName()
	p.MachineName = mcnName
	c.Client.MachineName = mcnName
	c.plugin = p

	return c, nil
}

func (c *RPCClientDriver) MarshalJSON() ([]byte, error) {
	return c.GetConfigRaw()
}

func (c *RPCClientDriver) UnmarshalJSON(data []byte) error {
	return c.SetConfigRaw(data)
}

func (c *RPCClientDriver) close() error {
	c.heartbeatDoneCh <- true
	close(c.heartbeatDoneCh)

	log.Debugf("Making call to close driver server")

	if err := c.Client.Call(CloseMethod, struct{}{}, nil); err != nil {
		log.Debugf("Failed to make call to close driver server: %s", err)
	} else {
		log.Debugf("Successfully made call to close driver server")
	}

	log.Debugf("Making call to close connection to plugin binary")

	return c.plugin.Close()
}

// Helper method to make requests which take no arguments and return simply a
// string, e.g. "GetIP".
func (c *RPCClientDriver) rpcStringCall(method string) (string, error) {
	var info string

	if err := c.Client.Call(method, struct{}{}, &info); err != nil {
		return "", err
	}

	return info, nil
}

func (c *RPCClientDriver) SetConfigRaw(data []byte) error {
	return c.Client.Call(SetConfigRawMethod, data, nil)
}

func (c *RPCClientDriver) UpdateConfigRaw(data []byte) error {
	return c.Client.Call(UpdateConfigRawMethod, data, nil)
}

func (c *RPCClientDriver) GetConfigRaw() ([]byte, error) {
	var data []byte

	if err := c.Client.Call(GetConfigRawMethod, struct{}{}, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// DriverName returns the name of the driver
func (c *RPCClientDriver) DriverName() string {
	driverName, err := c.rpcStringCall(DriverNameMethod)
	if err != nil {
		log.Warnf("Error attempting call to get driver name: %s", err)
	}

	return driverName
}

// DriverVersion returns the name of the driver
func (c *RPCClientDriver) DriverVersion() string {
	driverVersion, err := c.rpcStringCall(DriverNameMethod)
	if err != nil {
		log.Warnf("Error attempting call to get driver version: %s", err)
	}

	return driverVersion
}

func (c *RPCClientDriver) GetMachineName() string {
	name, err := c.rpcStringCall(GetMachineNameMethod)
	if err != nil {
		log.Warnf("Error attempting call to get machine name: %s", err)
	}

	return name
}

func (c *RPCClientDriver) GetIP() (string, error) {
	return c.rpcStringCall(GetIPMethod)
}

func (c *RPCClientDriver) GetBundleName() (string, error) {
	return c.rpcStringCall(GetBundleNameMethod)
}

func (c *RPCClientDriver) GetState() (state.State, error) {
	var s state.State

	if err := c.Client.Call(GetStateMethod, struct{}{}, &s); err != nil {
		return state.Error, err
	}

	return s, nil
}

func (c *RPCClientDriver) PreCreateCheck() error {
	return c.Client.Call(PreCreateCheckMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) Create() error {
	return c.Client.Call(CreateMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) Remove() error {
	return c.Client.Call(RemoveMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) Start() error {
	return c.Client.Call(StartMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) Stop() error {
	return c.Client.Call(StopMethod, struct{}{}, nil)
}

func (c *RPCClientDriver) Kill() error {
	return c.Client.Call(KillMethod, struct{}{}, nil)
}
