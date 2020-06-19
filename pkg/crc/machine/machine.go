package machine

import (
	"fmt"
	"runtime"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/libmachine2"
	"github.com/code-ready/crc/pkg/crc/network"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	crcos "github.com/code-ready/crc/pkg/os"

	// cluster services
	"github.com/code-ready/crc/pkg/crc/oc"
	// machine related imports
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/machine/libmachine/drivers"
	"github.com/code-ready/machine/libmachine/log"
	"github.com/code-ready/machine/libmachine/ssh"
	"github.com/code-ready/machine/libmachine/state"
)

func init() {
	// Force using the golang SSH implementation for windows
	if runtime.GOOS == crcos.WINDOWS.String() {
		ssh.SetDefaultClient(ssh.Native)
	}
}

func IsRunning(st state.State) bool {
	return st == state.Running
}

func Stop(stopConfig StopConfig) (StopResult, error) {
	defer unsetMachineLogging()

	result := &StopResult{Name: stopConfig.Name}
	// Set libmachine logging
	err := setMachineLogging(stopConfig.Debug)
	if err != nil {
		return *result, errors.New(err.Error())
	}

	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(stopConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	defer libMachineAPIClient.Close()

	result.State, _ = host.Driver.GetState()

	if err := host.Stop(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	result.Success = true
	return *result, nil
}

func PowerOff(powerOff PowerOffConfig) (PowerOffResult, error) {
	result := &PowerOffResult{Name: powerOff.Name}

	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(powerOff.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	defer libMachineAPIClient.Close()

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

	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(deleteConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	defer libMachineAPIClient.Close()

	if err := host.Driver.Remove(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	if err := libMachineAPIClient.Remove(deleteConfig.Name); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	return *result, nil
}

func Ip(ipConfig IpConfig) (IpResult, error) {
	result := &IpResult{Name: ipConfig.Name, Success: true}

	err := setMachineLogging(ipConfig.Debug)
	if err != nil {
		return *result, err
	}

	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(ipConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	defer libMachineAPIClient.Close()
	if result.IP, err = host.Driver.GetIP(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	return *result, nil
}

func Status(statusConfig ClusterStatusConfig) (ClusterStatusResult, error) {
	result := &ClusterStatusResult{Name: statusConfig.Name, Success: true}
	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	defer libMachineAPIClient.Close()

	_, err := libMachineAPIClient.Exists(statusConfig.Name)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	openshiftStatus := "Stopped"
	var diskUse int64
	var diskSize int64

	host, err := libMachineAPIClient.Load(statusConfig.Name)
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

	if IsRunning(vmStatus) {
		_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error loading bundle metadata: %v", err)
		}
		proxyConfig, err := getProxyConfig(crcBundleMetadata.ClusterInfo.BaseDomain)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error getting proxy configuration: %v", err)
		}
		proxyConfig.ApplyToEnvironment()

		// check if all the clusteroperators are running
		ocConfig := oc.UseOCWithConfig(statusConfig.Name)
		operatorsStatus, err := oc.GetClusterOperatorStatus(ocConfig)
		if err != nil {
			openshiftStatus = "Not Reachable"
			logging.Debug(err.Error())
		}
		switch {
		case operatorsStatus.Available:
			openshiftVersion := "4.x"
			if crcBundleMetadata.GetOpenshiftVersion() != "" {
				openshiftVersion = crcBundleMetadata.GetOpenshiftVersion()
			}
			openshiftStatus = fmt.Sprintf("Running (v%s)", openshiftVersion)
		case operatorsStatus.Degraded:
			openshiftStatus = "Degraded"
		case operatorsStatus.Progressing:
			openshiftStatus = "Starting"
		}
		sshRunner := crcssh.CreateRunner(host.Driver)
		diskSize, diskUse, err = cluster.GetRootPartitionUsage(sshRunner)
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

func MachineExists(name string) (bool, error) {
	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	defer libMachineAPIClient.Close()
	exists, err := libMachineAPIClient.Exists(name)
	if err != nil {
		return false, fmt.Errorf("Error checking if the host exists: %s", err)
	}
	return exists, nil
}

// Return proxy config if VM is present
func GetProxyConfig(machineName string) (*network.ProxyConfig, error) {
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(machineName)

	if err != nil {
		return nil, errors.New(err.Error())
	}

	_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Newf("Error loading bundle metadata: %v", err)
	}

	var clusterConfig ClusterConfig

	err = fillClusterConfig(crcBundleMetadata, &clusterConfig)
	if err != nil {
		return nil, errors.Newf("Error loading cluster configuration: %v", err)
	}

	return clusterConfig.ProxyConfig, nil
}

// Return console URL if the VM is present.
func GetConsoleURL(consoleConfig ConsoleConfig) (ConsoleResult, error) {
	result := &ConsoleResult{Success: true}
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	libMachineAPIClient := libmachine2.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(consoleConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}

	vmState, err := host.Driver.GetState()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.Newf("Error getting the state for host: %v", err)
	}
	result.State = vmState

	_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error loading bundle metadata: %v", err)
	}
	err = fillClusterConfig(crcBundleMetadata, &result.ClusterConfig)
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error loading cluster configuration: %v", err)
	}

	return *result, nil
}

func fillClusterConfig(bundleInfo *bundle.CrcBundleInfo, clusterConfig *ClusterConfig) error {
	kubeadminPassword, err := bundleInfo.GetKubeadminPassword()
	if err != nil {
		return fmt.Errorf("Error reading kubeadmin password from bundle %v", err)
	}

	proxyConfig, err := getProxyConfig(bundleInfo.ClusterInfo.BaseDomain)
	if err != nil {
		return err
	}
	*clusterConfig = ClusterConfig{
		KubeConfig:    bundleInfo.GetKubeConfigPath(),
		KubeAdminPass: kubeadminPassword,
		WebConsoleURL: constants.DefaultWebConsoleURL,
		ClusterAPI:    constants.DefaultAPIURL,
		ProxyConfig:   proxyConfig,
	}

	return nil
}

func getBundleMetadataFromDriver(driver drivers.Driver) (string, *bundle.CrcBundleInfo, error) {
	bundleName, err := driver.GetBundleName()
	/* FIXME: the bundleName == "" check can be removed when all machine
	* drivers have been rebuilt with
	* https://github.com/code-ready/machine/commit/edeebfe54d1ca3f46c1c0bfb86846e54baf23708
	 */
	if bundleName == "" || err != nil {
		err := fmt.Errorf("Error getting bundle name from CodeReady Containers instance, make sure you ran 'crc setup' and are using the latest bundle")
		return "", nil, err
	}
	metadata, err := bundle.GetCachedBundleInfo(bundleName)
	if err != nil {
		return "", nil, err
	}

	return bundleName, metadata, err
}

func setMachineLogging(logs bool) error {
	if !logs {
		log.SetDebug(true)
		logging.RemoveFileHook()
		logfile, err := logging.OpenLogFile(constants.LogFilePath)
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
	logging.SetupFileHook(constants.LogFilePath)
}

func getProxyConfig(baseDomainName string) (*network.ProxyConfig, error) {
	proxy, err := network.NewProxyConfig()
	if err != nil {
		return nil, err
	}
	if proxy.IsEnabled() {
		proxy.AddNoProxy(fmt.Sprintf(".%s", baseDomainName))
	}

	return proxy, nil
}
