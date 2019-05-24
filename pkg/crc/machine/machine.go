package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/libvirt"
	"github.com/code-ready/crc/pkg/crc/machine/virtualbox"
	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/host"
	"github.com/code-ready/machine/libmachine/log"
	"github.com/code-ready/machine/libmachine/state"
)

func init() {
}

func Start(startConfig StartConfig) (StartResult, error) {
	defer unsetMachineLogging()

	result := &StartResult{Name: startConfig.Name}

	// Set libmachine logging
	err := setMachineLogging(startConfig.Debug)
	if err != nil {
		return *result, err
	}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	defer libMachineAPIClient.Close()

	machineConfig := config.MachineConfig{
		Name:       startConfig.Name,
		BundlePath: startConfig.BundlePath,
		VMDriver:   startConfig.VMDriver,
		CPUs:       startConfig.CPUs,
		Memory:     startConfig.Memory,
	}

	logging.InfoF(" Extracting the Bundle tarball ...")
	crcBundleMetadata, extractedPath, err := bundle.GetCrcBundleInfo(machineConfig)
	if err != nil {
		logging.ErrorF("Error to get bundle Metadata %v", err)
		result.Error = err.Error()
		return *result, err
	}

	// Retrieve metadata info
	diskPath := filepath.Join(extractedPath, crcBundleMetadata.Storage.DiskImages[0].Name)
	machineConfig.DiskPathURL = fmt.Sprintf("file://%s", diskPath)
	machineConfig.SSHKeyPath = filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.SSHPrivateKeyFile)

	// Get the content of kubeadmin-password file
	kubeadminPassword, err := ioutil.ReadFile(filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.KubeadminPasswordFile))
	if err != nil {
		logging.ErrorF("Error reading the %s file %v", filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.KubeadminPasswordFile), err)
		result.Error = err.Error()
		return *result, err
	}

	// Put ClusterInfo to StartResult config.
	clusterConfig := ClusterConfig{
		KubeConfig:    filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.KubeConfig),
		KubeAdminPass: string(kubeadminPassword),
		ClusterAPI:    constants.DefaultWebConsoleURL,
	}

	result.ClusterConfig = clusterConfig

	exists, err := existVM(libMachineAPIClient, machineConfig)
	if !exists {
		logging.InfoF(" Creating VM ...")

		host, err := createHost(libMachineAPIClient, machineConfig)
		if err != nil {
			logging.ErrorF("Error creating host: %s", err)
			result.Error = err.Error()
		}

		vmState, err := host.Driver.GetState()
		if err != nil {
			logging.ErrorF("Error getting the state for host: %s", err)
			result.Error = err.Error()
		}

		result.Status = vmState.String()
	} else {
		logging.InfoF(" Starting stopped VM ...")
		host, err := libMachineAPIClient.Load(machineConfig.Name)
		s, err := host.Driver.GetState()
		if err != nil {
			logging.ErrorF("Error getting the state for host: %s", err)
			result.Error = err.Error()
		}

		if s != state.Running {
			if err := host.Driver.Start(); err != nil {
				logging.ErrorF("Error starting stopped VM: %s", err)
				result.Error = err.Error()
			}
			if err := libMachineAPIClient.Save(host); err != nil {
				logging.ErrorF("Error saving state for VM: %s", err)
				result.Error = err.Error()
			}
		}

		vmState, err := host.Driver.GetState()
		if err != nil {
			logging.ErrorF("Error getting the state for host: %s", err)
			result.Error = err.Error()
		}

		result.Status = vmState.String()
	}

	logging.InfoF(" Waiting 3m0s for the openshift cluster to be started ...")
	time.Sleep(time.Minute * 3)
	logging.InfoF(" To access the cluster as the system:admin user when using 'oc', run 'export KUBECONFIG=%s'", result.ClusterConfig.KubeConfig)
	logging.InfoF(" Access the OpenShift web-console here: %s", result.ClusterConfig.ClusterAPI)
	logging.InfoF(" Login to the console with user: kubeadmin, password: %s", result.ClusterConfig.KubeAdminPass)

	return *result, err
}

func Stop(stopConfig StopConfig) (StopResult, error) {
	defer unsetMachineLogging()

	result := &StopResult{Name: stopConfig.Name}
	// Set libmachine logging
	err := setMachineLogging(stopConfig.Debug)
	if err != nil {
		return *result, err
	}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(stopConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	if err := host.Stop(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	result.Success = true
	return *result, nil
}

func Delete(deleteConfig DeleteConfig) (DeleteResult, error) {
	result := &DeleteResult{Name: deleteConfig.Name, Success: true}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(deleteConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, err
	}

	m := errors.MultiError{}
	m.Collect(host.Driver.Remove())
	m.Collect(libMachineAPIClient.Remove(deleteConfig.Name))

	if len(m.Errors) != 0 {
		result.Success = false
		result.Error = m.ToError().Error()
		return *result, m.ToError()
	}
	return *result, nil
}

func existVM(api libmachine.API, machineConfig config.MachineConfig) (bool, error) {
	exists, err := api.Exists(machineConfig.Name)
	if err != nil {
		return false, errors.NewF("Error checking if the host exists: %s", err)
	}
	return exists, nil
}

func createHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	driverOptions := getDriverOptions(machineConfig)
	jsonDriverConfig, err := json.Marshal(driverOptions)

	vm, err := api.NewHost(machineConfig.VMDriver, jsonDriverConfig)

	if err != nil {
		return nil, errors.NewF("Error creating new host: %s", err)
	}

	if err := api.Create(vm); err != nil {
		return nil, errors.NewF("Error creating the VM. %s", err)
	}

	return vm, nil
}

func getDriverOptions(machineConfig config.MachineConfig) interface{} {
	var driver interface{}

	// Supported drivers
	switch machineConfig.VMDriver {

	case "libvirt":
		driver = libvirt.CreateHost(machineConfig)
	case "virtualbox":
		driver = virtualbox.CreateHost(machineConfig)

	default:
		errors.ExitWithMessage(1, "Unsupported driver: %s", machineConfig.VMDriver)
	}

	return driver
}

func setMachineLogging(logs bool) error {
	if !logs {
		log.SetDebug(true)
		logging.CloseLogFile()
		logfile, err := logging.OpenLogfile()
		if err != nil {
			return err
		}
		log.SetOutWriter(logfile)
		log.SetErrWriter(logfile)
	} else {
		log.SetDebug(true)
	}
	return nil
}

func unsetMachineLogging() {
	logging.CloseLogFile()
	logging.InitLogrus(logging.LogLevel)
}
