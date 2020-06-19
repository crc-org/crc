package libmachine2

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/code-ready/crc/pkg/crc/constants"

	"github.com/code-ready/crc/pkg/crc/machine/config"

	"github.com/code-ready/machine/libmachine/auth"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/host"
	"github.com/code-ready/machine/libmachine/log"
	"github.com/code-ready/machine/libmachine/mcnerror"
	"github.com/code-ready/machine/libmachine/mcnutils"
	"github.com/code-ready/machine/libmachine/persist"
	"github.com/code-ready/machine/libmachine/ssh"
	"github.com/code-ready/machine/libmachine/state"
	"github.com/code-ready/machine/libmachine/version"
)

type API interface {
	io.Closer
	NewHost(driverName string, driverPath string, rawDriver []byte) (*host.Host, error)
	Create(h *host.Host) error
	persist.Store
	GetMachinesDir() string
}

type Client struct {
	certsDir       string
	IsDebug        bool
	SSHClientType  ssh.ClientType
	GithubAPIToken string
	*persist.Filestore
}

func NewClient(storePath, certsDir string) *Client {
	return &Client{
		certsDir:      certsDir,
		IsDebug:       false,
		SSHClientType: ssh.External,
		Filestore:     persist.NewFilestore(storePath, certsDir, certsDir),
	}
}

func (api *Client) NewHost(machineConfig config.MachineConfig) (*host.Host, error) {
	driver := driver(machineConfig, &drivers.BaseDriver{
		MachineName: machineConfig.Name,
		StorePath:   constants.MachineBaseDir,
		SSHUser:     constants.DefaultSSHUser,
		BundleName:  machineConfig.BundleName,
	})

	return &host.Host{
		ConfigVersion: version.ConfigVersion,
		Name:          driver.GetMachineName(),
		Driver:        driver,
		DriverName:    driver.DriverName(),
		HostOptions: &host.Options{
			AuthOptions: &auth.Options{
				CertDir:          api.certsDir,
				CaCertPath:       filepath.Join(api.certsDir, "ca.pem"),
				CaPrivateKeyPath: filepath.Join(api.certsDir, "ca-key.pem"),
				ClientCertPath:   filepath.Join(api.certsDir, "cert.pem"),
				ClientKeyPath:    filepath.Join(api.certsDir, "key.pem"),
				ServerCertPath:   filepath.Join(api.GetMachinesDir(), "server.pem"),
				ServerKeyPath:    filepath.Join(api.GetMachinesDir(), "server-key.pem"),
			},
		},
	}, nil
}

func (api *Client) Load(name string) (*host.Host, error) {
	h, err := api.Filestore.Load(name)
	if err != nil {
		return nil, err
	}

	var machineConfig config.MachineConfig
	if err := json.Unmarshal(h.RawDriver, &machineConfig); err != nil {
		return nil, err
	}

	var baseDriver drivers.BaseDriver
	if err := json.Unmarshal(h.RawDriver, &baseDriver); err != nil {
		return nil, err
	}

	h.Driver = driver(machineConfig, &baseDriver)

	return h, nil
}

// Create is the wrapper method which covers all of the boilerplate around
// actually creating, provisioning, and persisting an instance in the store.
func (api *Client) Create(h *host.Host) error {
	log.Info("Running pre-create checks...")

	if err := h.Driver.PreCreateCheck(); err != nil {
		return mcnerror.ErrDuringPreCreate{
			Cause: err,
		}
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store before attempting creation: %s", err)
	}

	log.Info("Creating machine...")

	if err := api.performCreate(h); err != nil {
		return fmt.Errorf("Error creating machine: %s", err)
	}

	log.Debug("Machine successfully created")
	if err := api.SetExists(h.Name); err != nil {
		log.Debug("Failed to record VM existence")
	}

	return nil
}

func (api *Client) performCreate(h *host.Host) error {
	if err := h.Driver.Create(); err != nil {
		return fmt.Errorf("Error in driver during machine creation: %s", err)
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error saving host to store after attempting creation: %s", err)
	}

	// TODO: Not really a fan of just checking "none" or "ci-test" here.
	if h.Driver.DriverName() == "none" || h.Driver.DriverName() == "ci-test" {
		return nil
	}

	log.Info("Waiting for machine to be running, this may take a few minutes...")
	if err := mcnutils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
		return fmt.Errorf("Error waiting for machine to be running: %s", err)
	}

	log.Info("Machine is up and running!")
	return nil
}

func (api *Client) Close() error {
	return nil
}
