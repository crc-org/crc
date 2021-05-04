package libmachine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	log "github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/drivers/hyperv"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/code-ready/crc/pkg/libmachine/persist"
	"github.com/code-ready/machine/libmachine/drivers"
	rpcdriver "github.com/code-ready/machine/libmachine/drivers/rpc"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/pkg/errors"
)

type API interface {
	io.Closer
	NewHost(driverName string, driverPath string, rawDriver []byte) (*host.Host, error)
	Create(ctx context.Context, h *host.Host) error
	persist.Store
}

type Client struct {
	*persist.Filestore
	clientDriverFactory rpcdriver.RPCClientDriverFactory
}

func NewClient(storePath string) *Client {
	return &Client{
		Filestore:           persist.NewFilestore(storePath),
		clientDriverFactory: rpcdriver.NewRPCClientDriverFactory(),
	}
}

func (api *Client) NewHost(driverName string, driverPath string, rawDriver []byte) (*host.Host, error) {
	var driver drivers.Driver
	if driverName == "hyperv" {
		driver = hyperv.NewDriver("", "")
		if err := json.Unmarshal(rawDriver, &driver); err != nil {
			return nil, err
		}
	} else {
		var err error
		driver, err = api.clientDriverFactory.NewRPCClientDriver(driverName, driverPath, rawDriver)
		if err != nil {
			return nil, err
		}
	}

	return &host.Host{
		ConfigVersion: host.Version,
		Name:          driver.GetMachineName(),
		Driver:        driver,
		DriverName:    driver.DriverName(),
		DriverPath:    driverPath,
		RawDriver:     rawDriver,
	}, nil
}

func (api *Client) Load(name string) (*host.Host, error) {
	h, err := api.Filestore.Load(name)
	if err != nil {
		return nil, err
	}

	if h.DriverName == "hyperv" {
		driver := hyperv.NewDriver("", "")
		if err := json.Unmarshal(h.RawDriver, &driver); err != nil {
			return nil, err
		}
		h.Driver = driver
		return h, nil
	}

	d, err := api.clientDriverFactory.NewRPCClientDriver(h.DriverName, h.DriverPath, h.RawDriver)
	if err != nil {
		return nil, err
	}
	h.Driver = d
	return h, nil
}

// Create is the wrapper method which covers all of the boilerplate around
// actually creating, provisioning, and persisting an instance in the store.
func (api *Client) Create(ctx context.Context, h *host.Host) error {
	log.Debug("Running pre-create checks...")

	if err := h.Driver.PreCreateCheck(); err != nil {
		return errors.Wrap(err, "error with pre-create check")
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store before attempting creation: %s", err)
	}

	log.Debug("Creating machine...")

	if err := h.Driver.Create(); err != nil {
		return fmt.Errorf("Error in driver during machine creation: %s", err)
	}

	if err := h.Driver.Start(); err != nil {
		return fmt.Errorf("Error in driver during machine start: %s", err)
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store after attempting creation: %s", err)
	}

	log.Debug("Waiting for machine to be running, this may take a few minutes...")
	if err := crcerrors.RetryAfterWithContext(ctx, 3*time.Minute, host.MachineInState(h.Driver, state.Running), 3*time.Second); err != nil {
		return fmt.Errorf("Error waiting for machine to be running: %s", err)
	}

	log.Debug("Machine is up and running!")

	if err := api.SetExists(h.Name); err != nil {
		log.Debug("Failed to record VM existence")
	}

	return nil
}

func (api *Client) Close() error {
	return api.clientDriverFactory.Close()
}
