package machine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"time"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
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
	"github.com/code-ready/machine/libmachine/ssh"
	"github.com/code-ready/machine/libmachine/state"
)

func init() {
	// Force using the golang SSH implementation for windows
	if runtime.GOOS == crcos.WINDOWS.String() {
		ssh.SetDefaultClient(ssh.Native)
	}
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

func getCrcBundleInfo(bundlePath string) (*bundle.CrcBundleInfo, error) {
	bundleName := filepath.Base(bundlePath)
	bundleInfo, err := bundle.GetCachedBundleInfo(bundleName)
	if err == nil {
		logging.Infof("Loading bundle: %s ...", bundleName)
		return bundleInfo, nil
	}
	logging.Infof("Extracting bundle: %s ...", bundleName)
	return bundle.Extract(bundlePath)
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

func IsRunning(st state.State) bool {
	return st == state.Running
}

func Start(startConfig StartConfig) (StartResult, error) {
	var crcBundleMetadata *bundle.CrcBundleInfo
	defer unsetMachineLogging()

	result := &StartResult{Name: startConfig.Name}

	// Set libmachine logging
	err := setMachineLogging(startConfig.Debug)
	if err != nil {
		return *result, errors.New(err.Error())
	}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	defer libMachineAPIClient.Close()

	// Pre-VM start
	var privateKeyPath string
	var pullSecret string
	driverInfo, _ := getDriverInfo(startConfig.VMDriver)
	exists, err := MachineExists(startConfig.Name)
	if !exists {
		machineConfig := config.MachineConfig{
			Name:       startConfig.Name,
			BundleName: filepath.Base(startConfig.BundlePath),
			VMDriver:   startConfig.VMDriver,
			CPUs:       startConfig.CPUs,
			Memory:     startConfig.Memory,
		}

		pullSecret, err = startConfig.GetPullSecret()
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Failed to get pull secret: %v", err)
		}

		crcBundleMetadata, err = getCrcBundleInfo(startConfig.BundlePath)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error getting bundle metadata: %v", err)
		}

		openshiftVersion := crcBundleMetadata.GetOpenshiftVersion()
		if openshiftVersion == "" {
			logging.Infof("Creating VM...")
		} else {
			logging.Infof("Creating CodeReady Containers VM for OpenShift %s...", openshiftVersion)
		}

		// Retrieve metadata info
		diskPath := crcBundleMetadata.GetDiskImagePath()
		machineConfig.DiskPathURL = fmt.Sprintf("file://%s", filepath.ToSlash(diskPath))
		machineConfig.SSHKeyPath = crcBundleMetadata.GetSSHKeyPath()
		machineConfig.KernelCmdLine = crcBundleMetadata.Nodes[0].KernelCmdLine
		machineConfig.Initramfs = crcBundleMetadata.GetInitramfsPath()
		machineConfig.Kernel = crcBundleMetadata.GetKernelPath()

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
		privateKeyPath = crcBundleMetadata.GetSSHKeyPath()
	} else {
		host, err := libMachineAPIClient.Load(startConfig.Name)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error loading host: %v", err)
		}

		var bundleName string
		bundleName, crcBundleMetadata, err = getBundleMetadataFromDriver(host.Driver)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error loading bundle metadata: %v", err)
		}
		if bundleName != filepath.Base(startConfig.BundlePath) {
			logging.Fatalf("Bundle '%s' was requested, but the existing VM is using '%s'",
				filepath.Base(startConfig.BundlePath), bundleName)
		}
		if host.Driver.DriverName() != startConfig.VMDriver {
			err := errors.Newf("VM driver '%s' was requested, but the existing VM is using '%s' instead",
				startConfig.VMDriver, host.Driver.DriverName())
			result.Error = err.Error()
			return *result, err
		}
		vmState, err := host.Driver.GetState()
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error getting the state for host: %v", err)
		}
		if IsRunning(vmState) {
			openshiftVersion := crcBundleMetadata.GetOpenshiftVersion()
			if openshiftVersion == "" {
				logging.Infof("A CodeReady Containers VM is already running")
			} else {
				logging.Infof("A CodeReady Containers VM for OpenShift %s is already running", openshiftVersion)
			}
			result.Status = vmState.String()
			return *result, nil
		} else {
			openshiftVersion := crcBundleMetadata.GetOpenshiftVersion()
			if openshiftVersion == "" {
				logging.Infof("Starting CodeReady Containers VM ...")
			} else {
				logging.Infof("Starting CodeReady Containers VM for OpenShift %s...", openshiftVersion)
			}
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
		privateKeyPath = constants.GetPrivateKeyPath()
	}

	err = fillClusterConfig(crcBundleMetadata, &result.ClusterConfig)
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("%s", err.Error())
	}

	// Post-VM start
	host, err := libMachineAPIClient.Load(startConfig.Name)
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error loading %s vm: %v", startConfig.Name, err)
	}
	sshRunner := crcssh.CreateRunnerWithPrivateKey(host.Driver, privateKeyPath)

	logging.Debug("Waiting until ssh is available")
	if err := cluster.WaitForSsh(sshRunner); err != nil {
		result.Error = err.Error()
		return *result, errors.New("Failed to connect to the CRC VM with SSH")
	}

	// Check the certs validity inside the vm
	needsCertsRenewal := false
	logging.Info("Verifying validity of the cluster certificates ...")
	certExpiryState, err := cluster.CheckCertsValidity(sshRunner)
	if err != nil {
		if certExpiryState == cluster.CertExpired {
			needsCertsRenewal = true
		} else {
			result.Error = err.Error()
			return *result, errors.New(err.Error())
		}
	}
	// Add nameserver to VM if provided by User
	if startConfig.NameServer != "" {
		if err = addNameServerToInstance(sshRunner, startConfig.NameServer); err != nil {
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
	determineHostIP := func() error {
		hostIP, err = network.DetermineHostIP(instanceIP)
		if err != nil {
			logging.Debugf("Error finding host IP (%v) - retrying", err)
			return &errors.RetriableError{Err: err}
		}
		return nil
	}

	if err := errors.RetryAfter(30, determineHostIP, 2*time.Second); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error determining host IP: %v", err)
	}

	// Create servicePostStartConfig for DNS checks and DNS start.
	servicePostStartConfig := services.ServicePostStartConfig{
		Name:       startConfig.Name,
		DriverName: host.Driver.DriverName(),
		// TODO: would prefer passing in a more generic type
		SSHRunner: sshRunner,
		IP:        instanceIP,
		HostIP:    hostIP,
		// TODO: should be more finegrained
		BundleMetadata: *crcBundleMetadata,
	}

	// Run the DNS server inside the VM
	if _, err := dns.RunPostStart(servicePostStartConfig); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error running post start: %v", err)
	}

	// Check DNS lookup before starting the kubelet
	if queryOutput, err := dns.CheckCRCLocalDNSReachable(servicePostStartConfig); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Failed internal DNS query: %v : %s", err, queryOutput)
	}
	logging.Infof("Check internal and public DNS query ...")

	if queryOutput, err := dns.CheckCRCPublicDNSReachable(servicePostStartConfig); err != nil {
		logging.Warnf("Failed public DNS query from the cluster: %v : %s", err, queryOutput)
	}

	// Additional steps to perform after newly created VM is up
	if !exists {
		if err := updateSSHKeyPair(sshRunner); err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error updating public key: %v", err)
		}
		// Copy Kubeconfig file from bundle extract path to machine directory.
		// In our case it would be ~/.crc/machines/crc/
		logging.Infof("Copying kubeconfig file to instance dir ...")
		kubeConfigFilePath := filepath.Join(constants.MachineInstanceDir, startConfig.Name, "kubeconfig")
		err := crcos.CopyFileContents(crcBundleMetadata.GetKubeConfigPath(),
			kubeConfigFilePath,
			0644)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error copying kubeconfig file  %v", err)
		}
	}

	if needsCertsRenewal {
		logging.Infof("Cluster TLS certificates have expired, renewing them...")
		err = RegenerateCertificates(sshRunner, startConfig.Name)
		if err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Failed to renew TLS certificates: %v", err)
		}
	}

	ocConfig := oc.UseOCWithConfig(startConfig.Name)
	if !exists {
		if err := configureCluster(ocConfig, sshRunner, crcBundleMetadata.ClusterInfo.BaseDomain, pullSecret, instanceIP); err != nil {
			result.Error = err.Error()
			return *result, errors.Newf("Error Setting cluster config: %s", err)
		}
	}

	// Start kubelet inside the VM
	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	kubeletStarted, err := sd.Start("kubelet")
	if err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error starting kubelet: %s", err)
	}
	if kubeletStarted {
		logging.Infof("Starting OpenShift cluster ... [waiting 3m]")
	}
	result.KubeletStarted = kubeletStarted

	time.Sleep(time.Minute * 3)

	// Approve the node certificate.
	if err := ocConfig.ApproveNodeCSR(); err != nil {
		result.Error = err.Error()
		return *result, errors.Newf("Error approving the node csr %v", err)
	}

	// If no error, return usage message
	if result.Error == "" {
		logging.Infof("")
		logging.Infof("To access the cluster, first set up your environment by following 'crc oc-env' instructions")
		logging.Infof("Then you can access it by running 'oc login -u developer -p developer %s'", result.ClusterConfig.ClusterAPI)
		logging.Infof("To login as an admin, run 'oc login -u kubeadmin -p %s %s'", result.ClusterConfig.KubeAdminPass, result.ClusterConfig.ClusterAPI)
		logging.Infof("")
		logging.Infof("You can now run 'crc console' and use these credentials to access the OpenShift web console")
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

func PowerOff(PowerOff PowerOffConfig) (PowerOffResult, error) {
	result := &PowerOffResult{Name: PowerOff.Name}

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(PowerOff.Name)

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

	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	host, err := libMachineAPIClient.Load(deleteConfig.Name)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	defer libMachineAPIClient.Close()

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
	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	_, err := libMachineAPIClient.Exists(statusConfig.Name)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		return *result, errors.New(err.Error())
	}
	defer libMachineAPIClient.Close()

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
		// check if all the clusteroperators are running
		ocConfig := oc.UseOCWithConfig(statusConfig.Name)
		operatorsRunning, err := oc.GetClusterOperatorStatus(ocConfig)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return *result, errors.New(err.Error())
		}
		if operatorsRunning {
			openshiftVersion := "4.x"
			_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
			if err != nil {
				logging.Debugf("Failed to load bundle metadata: %s", err.Error())
			} else if crcBundleMetadata.GetOpenshiftVersion() != "" {
				openshiftVersion = crcBundleMetadata.GetOpenshiftVersion()
			}
			openshiftStatus = fmt.Sprintf("Running (v%s)", openshiftVersion)
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
	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	defer libMachineAPIClient.Close()
	exists, err := libMachineAPIClient.Exists(name)
	if err != nil {
		return false, fmt.Errorf("Error checking if the host exists: %s", err)
	}
	return exists, nil
}

func createHost(api libmachine.API, driverPath string, machineConfig config.MachineConfig) (*host.Host, error) {
	driverOptions := getDriverOptions(machineConfig)
	jsonDriverConfig, err := json.Marshal(driverOptions)
	if err != nil {
		return nil, errors.New("Failed to marshal driver options")
	}

	vm, err := api.NewHost(machineConfig.VMDriver, driverPath, jsonDriverConfig)

	if err != nil {
		return nil, fmt.Errorf("Error creating new host: %s", err)
	}

	if err := api.Create(vm); err != nil {
		return nil, fmt.Errorf("Error creating the VM: %s", err)
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

func addNameServerToInstance(sshRunner *crcssh.SSHRunner, ns string) error {
	nameserver := network.NameServer{IPAddress: ns}
	nameservers := []network.NameServer{nameserver}
	exist, err := network.HasGivenNameserversConfigured(sshRunner, nameserver)
	if err != nil {
		return err
	}
	if !exist {
		logging.Infof("Adding %s as nameserver to the instance ...", nameserver.IPAddress)
		network.AddNameserversToInstance(sshRunner, nameservers)
	}
	return nil
}

// Return proxy config if VM is present
func GetProxyConfig(machineName string) (*network.ProxyConfig, error) {
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
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
	libMachineAPIClient := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
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

func updateSSHKeyPair(sshRunner *crcssh.SSHRunner) error {
	// Generate ssh key pair
	if err := ssh.GenerateSSHKey(constants.GetPrivateKeyPath()); err != nil {
		return fmt.Errorf("Error generating ssh key pair: %v", err)
	}

	// Read generated public key
	publicKey, err := ioutil.ReadFile(constants.GetPublicKeyPath())
	if err != nil {
		return err
	}
	cmd := fmt.Sprintf("echo '%s' > /home/core/.ssh/authorized_keys", publicKey)
	_, err = sshRunner.Run(cmd)
	if err != nil {
		return err
	}
	sshRunner.SetPrivateKeyPath(constants.GetPrivateKeyPath())

	return err
}

func configureCluster(ocConfig oc.OcConfig, sshRunner *crcssh.SSHRunner, baseDomainName, pullSecret, instanceIP string) (rerr error) {
	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	if _, err := sd.Start("kubelet"); err != nil {
		return fmt.Errorf("Error starting kubelet: %s", err)
	}

	defer func() {
		// Stop the kubelet service.
		if _, err := sd.Stop("kubelet"); err != nil {
			rerr = err
		}
		if err := cluster.StopAndRemovePodsInVM(sshRunner); err != nil {
			rerr = err
		}
	}()
	if err := configProxyForCluster(ocConfig, sshRunner, sd, baseDomainName, instanceIP); err != nil {
		return fmt.Errorf("Failed to configure proxy for cluster: %v", err)
	}

	logging.Info("Adding user's pull secret ...")
	if err := cluster.AddPullSecret(sshRunner, ocConfig, pullSecret); err != nil {
		return fmt.Errorf("Failed to update user pull secret or cluster ID: %v", err)
	}
	logging.Info("Updating cluster ID ...")
	if err := cluster.UpdateClusterID(ocConfig); err != nil {
		return fmt.Errorf("Failed to update cluster ID: %v", err)
	}

	return nil
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

func configProxyForCluster(ocConfig oc.OcConfig, sshRunner *crcssh.SSHRunner, sd *systemd.InstanceSystemdCommander,
	baseDomainName, instanceIP string) (err error) {
	proxy, err := getProxyConfig(baseDomainName)
	if err != nil {
		return err
	}
	if proxy.IsEnabled() {
		defer func() {
			// Restart the crio service
			if proxy.IsEnabled() {
				// Restart reload the daemon and then restart the service
				// So no need to explicit reload the daemon.
				if _, ferr := sd.Restart("crio"); ferr != nil {
					err = ferr
				}
			}
		}()
		proxy.ApplyToEnvironment()

		logging.Info("Adding proxy configuration to the cluster ...")
		proxy.AddNoProxy(instanceIP)
		if err := cluster.AddProxyConfigToCluster(ocConfig, proxy); err != nil {
			return err
		}

		logging.Info("Adding proxy configuration to kubelet and crio service ...")
		if err := cluster.AddProxyToKubeletAndCriO(sshRunner, proxy); err != nil {
			return err
		}
	}
	return nil
}
