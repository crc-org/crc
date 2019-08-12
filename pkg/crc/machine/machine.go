package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/pullsecret"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"

	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/systemd"
	crcos "github.com/code-ready/crc/pkg/os"

	// cluster services
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/services"
	"github.com/code-ready/crc/pkg/crc/services/dns"

	// machine related imports
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/config"

	"github.com/code-ready/machine/libmachine"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/host"
	"github.com/code-ready/machine/libmachine/log"
	"github.com/code-ready/machine/libmachine/state"

	// cluster related import
	"github.com/code-ready/crc/pkg/crc/cluster"
)

func Start(startConfig StartConfig) (StartResult, error) {
	defer unsetMachineLogging()

	result := &StartResult{Name: startConfig.Name}

	// Set libmachine logging
	err := setMachineLogging(startConfig.Debug)
	if err != nil {
		return *result, errors.New(err.Error())
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

	logging.Infof("Extracting the %s Bundle tarball ...", filepath.Base(machineConfig.BundlePath))
	crcBundleMetadata, extractedPath, err := bundle.GetCrcBundleInfo(machineConfig)
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error to get bundle Metadata %v", err)
	}

	// Retrieve metadata info
	diskPath := filepath.Join(extractedPath, crcBundleMetadata.Storage.DiskImages[0].Name)
	// TODO: QnD Windows workaround hack
	machineConfig.DiskPathURL = fmt.Sprintf("file://%s", strings.Replace(diskPath, "\\", "/", -1))

	machineConfig.SSHKeyPath = filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.SSHPrivateKeyFile)
	machineConfig.KernelCmdLine = crcBundleMetadata.Nodes[0].KernelCmdLine
	machineConfig.Initramfs = filepath.Join(extractedPath, crcBundleMetadata.Nodes[0].Initramfs)
	machineConfig.Kernel = filepath.Join(extractedPath, crcBundleMetadata.Nodes[0].Kernel)

	// Get the content of kubeadmin-password file
	kubeadminPassword, err := ioutil.ReadFile(filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.KubeadminPasswordFile))
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error reading the %s file %v", filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.KubeadminPasswordFile), err)
	}

	// Put ClusterInfo to StartResult config.
	clusterConfig := ClusterConfig{
		KubeConfig:    filepath.Join(extractedPath, crcBundleMetadata.ClusterInfo.KubeConfig),
		KubeAdminPass: string(kubeadminPassword),
		WebConsoleURL: constants.DefaultWebConsoleURL,
		ClusterAPI:    constants.DefaultAPIURL,
	}

	result.ClusterConfig = clusterConfig

	// Pre-VM start
	driverInfo, _ := getDriverInfo(startConfig.VMDriver)
	exists, err := existVM(libMachineAPIClient, machineConfig)
	if !exists {
		logging.Infof("Creating VM ...")

		host, err := createHost(libMachineAPIClient, driverInfo.DriverPath, machineConfig)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error creating host: %v", err)
		}

		vmState, err := host.Driver.GetState()
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error getting the state for host: %v", err)
		}

		result.Status = vmState.String()
	} else {
		host, err := libMachineAPIClient.Load(machineConfig.Name)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error loading host: %v", err)
		}

		if host.Driver.DriverName() != startConfig.VMDriver {
			err := errors.Newf("VM driver '%s' was requested, but loaded VM is using '%s' instead",
				startConfig.VMDriver, host.Driver.DriverName())
			result.Error = err.Error()
			return *result, err
		}
		vmState, err := host.Driver.GetState()
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error getting the state for host: %v", err)
		}
		if vmState == state.Running {
			result.Status = vmState.String()
			return *result, nil
		}

		if vmState != state.Running {
			logging.Infof("Starting stopped VM ...")
			if err := host.Driver.Start(); err != nil {
				result.Error = err.Error()
				return *result, errors.Newf("Error starting stopped VM: %v", err)
			}
			if err := libMachineAPIClient.Save(host); err != nil {
				result.Error = err.Error()
				return *result, errors.Newf("Error saving state for VM: %v", err)
			}
		}

		vmState, err = host.Driver.GetState()
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error getting the state: %v", err)
		}

		result.Status = vmState.String()
	}

	// Post-VM start
	host, err := libMachineAPIClient.Load(machineConfig.Name)
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error loading %s vm: %v", machineConfig.Name, err)
	}

	logging.Debug("Waiting until ssh is available")
	if err := cluster.WaitForSsh(host.Driver); err != nil {
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	// Check the certs validity inside the vm
	logging.Info("Verifying validity of the cluster certificates ...")
	if err := cluster.CheckCertsValidity(host.Driver); err != nil {
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	// Add nameserver to VM if provided by User
	if startConfig.NameServer != "" {
		if addNameServerToInstance(host.Driver, startConfig.NameServer); err != nil {
			result.Error = err.Error()
			return *result, errors.New(err.Error())
		}
	}

	instanceIP, err := host.Driver.GetIP()
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error getting the IP: %v", err)
	}

	var hostIP string
	determinHostIP := func() error {
		hostIP, err = network.DetermineHostIP(instanceIP)
		if err != nil {
			logging.Debugf("Error finding host IP (%v) - retrying", err)
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	if err := errors.RetryAfter(30, determinHostIP, 2*time.Second); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error determining host IP: %v", err)
	}

	// Create servicePostStartConfig for dns checks and dns start.
	servicePostStartConfig := services.ServicePostStartConfig{
		Name: startConfig.Name,
		// TODO: would prefer passing in a more generic type
		Driver: host.Driver,
		IP:     instanceIP,
		HostIP: hostIP,
		// TODO: should be more finegrained
		BundleMetadata: *crcBundleMetadata,
	}

	// If driver need dns service then start it
	if driverInfo.UseDNSService {
		if _, err := dns.RunPostStart(servicePostStartConfig); err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error running post start: %v", err)
		}
	}
	// Check DNS looksup before starting the kubelet
	if queryOutput, err := dns.CheckCRCLocalDNSReachable(servicePostStartConfig); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Failed internal dns query: %v : %s", err, queryOutput)
	}
	logging.Infof("Check internal and public dns query ...")

	if queryOutput, err := dns.CheckCRCPublicDNSReachable(servicePostStartConfig); err != nil {
		logging.Warnf("Failed Public dns query: %v : %s", err, queryOutput)
	}

	// Copy Kubeconfig file from bundle extract path to machine directory.
	// In our case it would be ~/machine/crc
	logging.Infof("Copying kubeconfig file to instance dir ...")
	kubeConfigFilePath := filepath.Join(constants.MachineInstanceDir, machineConfig.Name, "kubeconfig")
	if err := crcos.CopyFileContents(
		filepath.Join(extractedPath, "kubeconfig"),
		kubeConfigFilePath,
		0644); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error to copy kubeconfig content %v", err)
	}

	// On VM creation, we need to add the user pull secret and generate a cluster ID
	if !exists {
		// Update the user pull secret before kubelet start.
		logging.Info("Adding user's pull secret and cluster ID ...")
		if err := pullsecret.AddPullSecretAndClusterID(host.Driver, startConfig.PullSecret, kubeConfigFilePath); err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Failed to update user pull secret or cluster ID: %v", err)
		}
	}

	// Start kubelet inside the VM
	sd := systemd.NewInstanceSystemdCommander(host.Driver)
	kubeletStarted, err := sd.Start("kubelet")
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error starting kubelet: %s", err)
	}
	if kubeletStarted {
		logging.Infof("Starting OpenShift cluster ... [waiting 3m]")
	}
	result.KubeletStarted = kubeletStarted

	// If no error, return usage message
	if result.Error == "" {
		time.Sleep(time.Minute * 3)
		logging.Infof("To access the cluster using 'oc', run 'oc login -u kubeadmin -p %s %s'", result.ClusterConfig.KubeAdminPass, result.ClusterConfig.ClusterAPI)
		logging.Infof("Access the OpenShift web-console here: %s", result.ClusterConfig.WebConsoleURL)
		logging.Infof("Login to the console with user: kubeadmin, password: %s", result.ClusterConfig.KubeAdminPass)
	}

	// Approve the node certificate.
	ocConfig := oc.UseOCWithConfig(machineConfig.Name)
	if err := oc.ApproveNodeCSR(ocConfig); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error approving the node csr %v", err)
	}

	return *result, err
}

func Stop(stopConfig StopConfig) (StopResult, error) {
	defer unsetMachineLogging()

	result := &StopResult{Name: stopConfig.Name}
	// Set libmachine logging
	err := setMachineLogging(stopConfig.Debug)
	if err != nil {
		return *result, errors.New(err.Error())
	}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(stopConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	result.State, _ = host.Driver.GetState()

	if err := host.Stop(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	result.Success = true
	return *result, nil
}

func PowerOff(PowerOff PowerOffConfig) (PowerOffResult, error) {
	result := &PowerOffResult{Name: PowerOff.Name}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(PowerOff.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	if err := host.Kill(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
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
		return *result, errors.New(err.Error())
	}

	m := errors.MultiError{}
	m.Collect(host.Driver.Remove())
	m.Collect(libMachineAPIClient.Remove(deleteConfig.Name))

	if len(m.Errors) != 0 {
		result.Success = false
		result.Error = m.ToError().Error()
		return *result, errors.New(m.ToError().Error())
	}
	return *result, nil
}

func Ip(ipConfig IpConfig) (IpResult, error) {
	result := &IpResult{Name: ipConfig.Name, Success: true}

	err := setMachineLogging(ipConfig.Debug)
	if err != nil {
		return *result, err
	}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(ipConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	if result.IP, err = host.Driver.GetIP(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	return *result, nil
}

func Status(statusConfig ClusterStatusConfig) (ClusterStatusResult, error) {
	result := &ClusterStatusResult{Name: statusConfig.Name, Success: true}
	api := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	_, err := api.Exists(statusConfig.Name)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	openshiftStatus := "Stopped"
	var diskUse int64
	var diskSize int64

	host, err := api.Load(statusConfig.Name)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	vmStatus, err := host.Driver.GetState()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	if vmStatus == state.Running {
		// check if all the clusteroperators are running
		ocConfig := oc.UseOCWithConfig(statusConfig.Name)
		operatorsRunning, err := oc.GetClusterOperatorStatus(ocConfig)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return *result, errors.New(err.Error())
		}
		if operatorsRunning {
			// TODO:get openshift version as well and add to status
			openshiftStatus = fmt.Sprintf("Running (v4.x)")
		}
		diskSize, diskUse, err = cluster.GetDiskUsage(host.Driver, "/dev/vda3")
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return *result, errors.New(err.Error())
		}
	}
	result.OpenshiftStatus = openshiftStatus
	result.DiskUse = diskUse
	result.DiskSize = diskSize
	result.CrcStatus = vmStatus.String()
	return *result, nil
}

func existVM(api libmachine.API, machineConfig config.MachineConfig) (bool, error) {
	exists, err := api.Exists(machineConfig.Name)
	if err != nil {
		return false, errors.Newf("Error checking if the host exists: %s", err)
	}
	return exists, nil
}

func createHost(api libmachine.API, driverPath string, machineConfig config.MachineConfig) (*host.Host, error) {
	driverOptions := getDriverOptions(machineConfig)
	jsonDriverConfig, err := json.Marshal(driverOptions)

	vm, err := api.NewHost(machineConfig.VMDriver, driverPath, jsonDriverConfig)

	if err != nil {
		return nil, errors.Newf("Error creating new host: %s", err)
	}

	if err := api.Create(vm); err != nil {
		return nil, errors.Newf("Error creating the VM. %s", err)
	}

	return vm, nil
}

func setMachineLogging(logs bool) error {
	if !logs {
		log.SetDebug(true)
		logging.RemoveFileHook()
		logfile, err := logging.OpenLogFile()
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
	logging.SetupFileHook()
}

func addNameServerToInstance(driver drivers.Driver, ns string) error {
	nameserver := network.NameServer{IPAddress: ns}
	nameservers := []network.NameServer{nameserver}
	exist, err := network.HasGivenNameserversConfigured(driver, nameserver)
	if err != nil {
		return err
	}
	if !exist {
		logging.Infof("Adding %s as nameserver to Instance ...", nameserver.IPAddress)
		network.AddNameserversToInstance(driver, nameservers)
	}
	return nil
}

// Return console URL if the VM is present.
func GetConsoleURL(consoleConfig ConsoleConfig) (ConsoleResult, error) {
	result := &ConsoleResult{Success: true}
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	_, err := libMachineAPIClient.Load(consoleConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	result.URL = constants.DefaultWebConsoleURL
	return *result, nil
}
