package machine

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/crc/systemd/states"
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

func getClusterConfig(bundleInfo *bundle.CrcBundleInfo) (*ClusterConfig, error) {
	kubeadminPassword, err := bundleInfo.GetKubeadminPassword()
	if err != nil {
		return nil, fmt.Errorf("Error reading kubeadmin password from bundle %v", err)
	}
	proxyConfig, err := getProxyConfig(bundleInfo.ClusterInfo.BaseDomain)
	if err != nil {
		return nil, err
	}
	clusterCACert, err := certificateAuthority(bundleInfo.GetKubeConfigPath())
	if err != nil {
		return nil, err
	}
	return &ClusterConfig{
		ClusterCACert: base64.StdEncoding.EncodeToString(clusterCACert),
		KubeConfig:    bundleInfo.GetKubeConfigPath(),
		KubeAdminPass: kubeadminPassword,
		WebConsoleURL: constants.DefaultWebConsoleURL,
		ClusterAPI:    constants.DefaultAPIURL,
		ProxyConfig:   proxyConfig,
	}, nil
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
	if err != nil {
		err := fmt.Errorf("Error getting bundle name from CodeReady Containers instance, make sure you ran 'crc setup' and are using the latest bundle")
		return "", nil, err
	}
	metadata, err := bundle.GetCachedBundleInfo(bundleName)
	if err != nil {
		return "", nil, err
	}

	return bundleName, metadata, err
}

func createLibMachineClient(debug bool) (*libmachine.Client, func(), error) {
	err := setMachineLogging(debug)
	if err != nil {
		return nil, func() {
			unsetMachineLogging()
		}, err
	}
	client := libmachine.NewClient(constants.MachineBaseDir, constants.MachineCertsDir)
	return client, func() {
		client.Close()
		unsetMachineLogging()
	}, nil
}

func (client *client) updateVMConfig(startConfig StartConfig, api libmachine.API, host *host.Host) error {
	/* Memory */
	logging.Debugf("Updating CRC VM configuration")
	if err := setMemory(host, startConfig.Memory); err != nil {
		logging.Debugf("Failed to update CRC VM configuration: %v", err)
		if err != drivers.ErrNotImplemented {
			return err
		}
	}
	if err := setVcpus(host, startConfig.CPUs); err != nil {
		logging.Debugf("Failed to update CRC VM configuration: %v", err)
		if err != drivers.ErrNotImplemented {
			return err
		}
	}
	if err := api.Save(host); err != nil {
		return err
	}

	return nil
}

func (client *client) Start(startConfig StartConfig) (StartResult, error) {
	var crcBundleMetadata *bundle.CrcBundleInfo

	libMachineAPIClient, cleanup, err := createLibMachineClient(startConfig.Debug)
	defer cleanup()
	if err != nil {
		return startError(startConfig.Name, "Cannot initialize libmachine", err)
	}

	// Pre-VM start
	var host *host.Host
	exists, err := client.Exists(startConfig.Name)
	if err != nil {
		return startError(startConfig.Name, "Cannot determine if VM exists", err)
	}
	if !exists {
		// Ask early for pull secret if it hasn't been requested yet
		_, err = startConfig.PullSecret.Value()
		if err != nil {
			return startError(startConfig.Name, "Failed to ask for pull secret", err)
		}

		machineConfig := config.MachineConfig{
			Name:       startConfig.Name,
			BundleName: filepath.Base(startConfig.BundlePath),
			CPUs:       startConfig.CPUs,
			Memory:     startConfig.Memory,
		}

		crcBundleMetadata, err = getCrcBundleInfo(startConfig.BundlePath)
		if err != nil {
			return startError(startConfig.Name, "Error getting bundle metadata", err)
		}

		logging.Infof("Checking size of the disk image %s ...", crcBundleMetadata.GetDiskImagePath())
		if err := crcBundleMetadata.CheckDiskImageSize(); err != nil {
			return startError(startConfig.Name, fmt.Sprintf("Invalid bundle disk image '%s'", crcBundleMetadata.GetDiskImagePath()), err)
		}

		logging.Infof("Creating CodeReady Containers VM for OpenShift %s...", crcBundleMetadata.GetOpenshiftVersion())

		// Retrieve metadata info
		machineConfig.ImageSourcePath = crcBundleMetadata.GetDiskImagePath()
		machineConfig.ImageFormat = crcBundleMetadata.Storage.DiskImages[0].Format
		machineConfig.SSHKeyPath = crcBundleMetadata.GetSSHKeyPath()
		machineConfig.KernelCmdLine = crcBundleMetadata.Nodes[0].KernelCmdLine
		machineConfig.Initramfs = crcBundleMetadata.GetInitramfsPath()
		machineConfig.Kernel = crcBundleMetadata.GetKernelPath()

		host, err = createHost(libMachineAPIClient, machineConfig)
		if err != nil {
			return startError(startConfig.Name, "Error creating machine", err)
		}
	} else {
		host, err = libMachineAPIClient.Load(startConfig.Name)
		if err != nil {
			return startError(startConfig.Name, "Error loading machine", err)
		}

		var bundleName string
		bundleName, crcBundleMetadata, err = getBundleMetadataFromDriver(host.Driver)
		if err != nil {
			return startError(startConfig.Name, "Error loading bundle metadata", err)
		}
		if bundleName != filepath.Base(startConfig.BundlePath) {
			logging.Debugf("Bundle '%s' was requested, but the existing VM is using '%s'",
				filepath.Base(startConfig.BundlePath), bundleName)
			return startError(
				startConfig.Name,
				"Invalid bundle",
				fmt.Errorf("Bundle '%s' was requested, but the existing VM is using '%s'",
					filepath.Base(startConfig.BundlePath),
					bundleName))
		}
		vmState, err := host.Driver.GetState()
		if err != nil {
			return startError(startConfig.Name, "Error getting the machine state", err)
		}
		if vmState == state.Running {
			logging.Infof("A CodeReady Containers VM for OpenShift %s is already running", crcBundleMetadata.GetOpenshiftVersion())
			clusterConfig, err := getClusterConfig(crcBundleMetadata)
			if err != nil {
				return startError(startConfig.Name, "Cannot create cluster configuration", err)
			}
			return StartResult{
				Name:           startConfig.Name,
				Status:         vmState.String(),
				ClusterConfig:  *clusterConfig,
				KubeletStarted: true,
			}, nil
		}

		logging.Infof("Starting CodeReady Containers VM for OpenShift %s...", crcBundleMetadata.GetOpenshiftVersion())

		if err := client.updateVMConfig(startConfig, libMachineAPIClient, host); err != nil {
			return startError(startConfig.Name, "Could not update CRC VM configuration", err)
		}

		if err := host.Driver.Start(); err != nil {
			return startError(startConfig.Name, "Error starting stopped VM", err)
		}
	}

	clusterConfig, err := getClusterConfig(crcBundleMetadata)
	if err != nil {
		return startError(startConfig.Name, "Cannot create cluster configuration", err)
	}

	// Post-VM start
	vmState, err := host.Driver.GetState()
	if err != nil {
		return startError(startConfig.Name, "Error getting the state", err)
	}
	if vmState != state.Running {
		return startError(startConfig.Name, "CodeReady Containers VM is not running", err)
	}

	instanceIP, err := host.Driver.GetIP()
	if err != nil {
		return startError(startConfig.Name, "Error getting the IP", err)
	}
	sshRunner := crcssh.CreateRunner(instanceIP, constants.DefaultSSHPort, crcBundleMetadata.GetSSHKeyPath(), constants.GetPrivateKeyPath())

	logging.Debug("Waiting until ssh is available")
	if err := cluster.WaitForSSH(sshRunner); err != nil {
		return startError(startConfig.Name, "Failed to connect to the CRC VM with SSH -- host might be unreachable", err)
	}
	logging.Info("CodeReady Containers VM is running")

	// Post VM start immediately update SSH key and copy kubeconfig to instance
	// dir and VM
	if err := updateSSHKeyAndCopyKubeconfig(sshRunner, startConfig, crcBundleMetadata); err != nil {
		return startError(startConfig.Name, "Error updating public key", err)
	}

	// Start network time synchronization if `CRC_DEBUG_ENABLE_STOP_NTP` is not set
	if stopNtp, _ := strconv.ParseBool(os.Getenv("CRC_DEBUG_ENABLE_STOP_NTP")); !stopNtp {
		logging.Info("Starting network time synchronization in CodeReady Containers VM")
		if _, err := sshRunner.Run("sudo timedatectl set-ntp on"); err != nil {
			return startError(startConfig.Name, "Failed to start network time synchronization", err)
		}
	}

	// Add nameserver to VM if provided by User
	if startConfig.NameServer != "" {
		if err = addNameServerToInstance(sshRunner, startConfig.NameServer); err != nil {
			return startError(startConfig.Name, "Failed to add nameserver to the VM", err)
		}
	}

	proxyConfig, err := getProxyConfig(crcBundleMetadata.ClusterInfo.BaseDomain)
	if err != nil {
		return startError(startConfig.Name, "Error getting proxy configuration", err)
	}
	proxyConfig.ApplyToEnvironment()

	// Create servicePostStartConfig for DNS checks and DNS start.
	servicePostStartConfig := services.ServicePostStartConfig{
		Name: startConfig.Name,
		// TODO: would prefer passing in a more generic type
		SSHRunner: sshRunner,
		IP:        instanceIP,
		// TODO: should be more finegrained
		BundleMetadata: *crcBundleMetadata,
	}

	// Run the DNS server inside the VM
	if _, err := dns.RunPostStart(servicePostStartConfig); err != nil {
		return startError(startConfig.Name, "Error running post start", err)
	}

	// Check DNS lookup before starting the kubelet
	if queryOutput, err := dns.CheckCRCLocalDNSReachable(servicePostStartConfig); err != nil {
		return startError(startConfig.Name, fmt.Sprintf("Failed internal DNS query: %s", queryOutput), err)
	}
	logging.Info("Check internal and public DNS query ...")

	if queryOutput, err := dns.CheckCRCPublicDNSReachable(servicePostStartConfig); err != nil {
		logging.Warnf("Failed public DNS query from the cluster: %v : %s", err, queryOutput)
	}

	// Check DNS lookup from host to VM
	logging.Info("Check DNS query from host ...")
	if err := network.CheckCRCLocalDNSReachableFromHost(crcBundleMetadata, instanceIP); err != nil {
		return startError(startConfig.Name, "Failed to query DNS from host", err)
	}

	if err := cluster.EnsurePullSecretPresentOnInstanceDisk(sshRunner, startConfig.PullSecret); err != nil {
		return startError(startConfig.Name, "Failed to update VM pull secret", err)
	}

	// Check the certs validity inside the vm
	logging.Info("Verifying validity of the kubelet certificates ...")
	clientExpired, serverExpired, err := cluster.CheckCertsValidity(sshRunner)
	if err != nil {
		return startError(startConfig.Name, "Failed to check certificate validity", err)
	}

	logging.Info("Starting OpenShift kubelet service")
	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	if err := sd.Start("kubelet"); err != nil {
		return startError(startConfig.Name, "Error starting kubelet", err)
	}

	ocConfig := oc.UseOCWithSSH(sshRunner)

	if err := cluster.WaitAndApprovePendingCSRs(ocConfig, clientExpired, serverExpired); err != nil {
		logBundleDate(crcBundleMetadata)
		return startError(startConfig.Name, "Failed to renew TLS certificates: please check if a newer CodeReady Containers release is available", err)
	}

	if !exists {
		logging.Info("Configuring cluster for first start")
		if err := configProxyForCluster(ocConfig, sshRunner, sd, proxyConfig, instanceIP); err != nil {
			return startError(startConfig.Name, "Error Setting cluster config", err)
		}
	}

	if err := cluster.EnsurePullSecretPresentInTheCluster(ocConfig, startConfig.PullSecret); err != nil {
		return startError(startConfig.Name, "Failed to update cluster pull secret", err)
	}

	if err := cluster.EnsureClusterIDIsNotEmpty(ocConfig); err != nil {
		return startError(startConfig.Name, "Failed to update cluster ID", err)
	}

	// Check if kubelet service is running inside the VM
	kubeletStatus, err := sd.Status("kubelet")
	if err != nil || kubeletStatus != states.Running {
		return startError(startConfig.Name, "kubelet service is not running", err)
	}

	// In Openshift 4.3, when cluster comes up, the following happens
	// 1. After the openshift-apiserver pod is started, its log contains multiple occurrences of `certificate has expired or is not yet valid`
	// 2. Initially there is no request-header's client-ca crt available to `extension-apiserver-authentication` configmap
	// 3. In the pod logs `missing content for CA bundle "client-ca::kube-system::extension-apiserver-authentication::requestheader-client-ca-file"`
	// 4. After ~1 min /etc/kubernetes/static-pod-resources/kube-apiserver-certs/configmaps/aggregator-client-ca/ca-bundle.crt is regenerated
	// 5. It is now also appear to `extension-apiserver-authentication` configmap as part of request-header's client-ca content
	// 6. Openshift-apiserver is able to load the CA which was regenerated
	// 7. Now apiserver pod log contains multiple occurrences of `error x509: certificate signed by unknown authority`
	// When the openshift-apiserver is in this state, the cluster is non functional.
	// A restart of the openshift-apiserver pod is enough to clear that error and get a working cluster.
	// This is a work-around while the root cause is being identified.
	// More info: https://bugzilla.redhat.com/show_bug.cgi?id=1795163
	logging.Debug("Waiting for update of client-ca request header ...")
	if err := cluster.WaitforRequestHeaderClientCaFile(ocConfig); err != nil {
		return startError(startConfig.Name, "Failed to wait for the client-ca request header update", err)
	}

	if err := cluster.DeleteOpenshiftAPIServerPods(ocConfig); err != nil {
		return startError(startConfig.Name, "Cannot delete OpenShift API Server pods", err)
	}

	logging.Info("Starting OpenShift cluster ... [waiting 3m]")

	time.Sleep(time.Minute * 3)

	logging.Info("Updating kubeconfig")
	if err := eventuallyWriteKubeconfig(ocConfig, instanceIP, clusterConfig); err != nil {
		log.Warnf("Cannot update kubeconfig: %v", err)
	}

	if proxyConfig.IsEnabled() {
		logging.Info("Waiting for the proxy configuration to be applied ...")
		waitForProxyPropagation(ocConfig, proxyConfig)
	}

	logging.Warn("The cluster might report a degraded or error state. This is expected since several operators have been disabled to lower the resource usage. For more information, please consult the documentation")
	return StartResult{
		Name:           startConfig.Name,
		KubeletStarted: true,
		ClusterConfig:  *clusterConfig,
		Status:         vmState.String(),
	}, err
}

func (*client) Stop(stopConfig StopConfig) (StopResult, error) {
	defer unsetMachineLogging()

	// Set libmachine logging
	err := setMachineLogging(stopConfig.Debug)
	if err != nil {
		return stopError(stopConfig.Name, "Cannot initialize logging", err)
	}

	libMachineAPIClient, cleanup, err := createLibMachineClient(stopConfig.Debug)
	defer cleanup()
	if err != nil {
		return stopError(stopConfig.Name, "Cannot initialize libmachine", err)
	}
	host, err := libMachineAPIClient.Load(stopConfig.Name)

	if err != nil {
		return stopError(stopConfig.Name, "Cannot load machine", err)
	}

	state, _ := host.Driver.GetState()

	logging.Info("Stopping the OpenShift cluster, this may take a few minutes...")
	if err := host.Stop(); err != nil {
		return stopError(stopConfig.Name, "Cannot stop machine", err)
	}

	return StopResult{
		Name:    stopConfig.Name,
		Success: true,
		State:   state,
	}, nil
}

func (*client) PowerOff(powerOff PowerOffConfig) (PowerOffResult, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return powerOffError(powerOff.Name, "Cannot initialize libmachine", err)
	}
	host, err := libMachineAPIClient.Load(powerOff.Name)

	if err != nil {
		return powerOffError(powerOff.Name, "Cannot load machine", err)
	}

	if err := host.Kill(); err != nil {
		return powerOffError(powerOff.Name, "Cannot kill machine", err)
	}

	return PowerOffResult{
		Name:    powerOff.Name,
		Success: true,
	}, nil
}

func (*client) Delete(deleteConfig DeleteConfig) (DeleteResult, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return deleteError(deleteConfig.Name, "Cannot initialize libmachine", err)
	}
	host, err := libMachineAPIClient.Load(deleteConfig.Name)

	if err != nil {
		return deleteError(deleteConfig.Name, "Cannot load machine", err)
	}

	if err := host.Driver.Remove(); err != nil {
		return deleteError(deleteConfig.Name, "Driver cannot remove machine", err)
	}

	if err := libMachineAPIClient.Remove(deleteConfig.Name); err != nil {
		return deleteError(deleteConfig.Name, "Cannot remove machine", err)
	}

	return DeleteResult{
		Name:    deleteConfig.Name,
		Success: true,
	}, nil
}

func (*client) IP(ipConfig IPConfig) (IPResult, error) {
	err := setMachineLogging(ipConfig.Debug)
	if err != nil {
		return ipError(ipConfig.Name, "Cannot initialize logging", err)
	}

	libMachineAPIClient, cleanup, err := createLibMachineClient(ipConfig.Debug)
	defer cleanup()
	if err != nil {
		return ipError(ipConfig.Name, "Cannot initialize libmachine", err)
	}
	host, err := libMachineAPIClient.Load(ipConfig.Name)

	if err != nil {
		return ipError(ipConfig.Name, "Cannot load machine", err)
	}
	ip, err := host.Driver.GetIP()
	if err != nil {
		return ipError(ipConfig.Name, "Cannot get IP", err)
	}
	return IPResult{
		Name:    ipConfig.Name,
		Success: true,
		IP:      ip,
	}, nil
}

func (*client) Status(statusConfig ClusterStatusConfig) (ClusterStatusResult, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return statusError(statusConfig.Name, "Cannot initialize libmachine", err)
	}

	_, err = libMachineAPIClient.Exists(statusConfig.Name)
	if err != nil {
		return statusError(statusConfig.Name, "Cannot check if machine exists", err)
	}

	openshiftStatus := "Stopped"
	openshiftVersion := ""
	var diskUse int64
	var diskSize int64

	host, err := libMachineAPIClient.Load(statusConfig.Name)
	if err != nil {
		return statusError(statusConfig.Name, "Cannot load machine", err)
	}
	vmStatus, err := host.Driver.GetState()
	if err != nil {
		return statusError(statusConfig.Name, "Cannot get machine state", err)
	}

	if vmStatus == state.Running {
		_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
		if err != nil {
			return statusError(statusConfig.Name, "Error loading bundle metadata", err)
		}
		proxyConfig, err := getProxyConfig(crcBundleMetadata.ClusterInfo.BaseDomain)
		if err != nil {
			return statusError(statusConfig.Name, "Error getting proxy configuration", err)
		}
		proxyConfig.ApplyToEnvironment()

		ip, err := host.Driver.GetIP()
		if err != nil {
			return statusError(statusConfig.Name, "Error getting ip", err)
		}
		sshRunner := crcssh.CreateRunner(ip, constants.DefaultSSHPort, constants.GetPrivateKeyPath())
		// check if all the clusteroperators are running
		ocConfig := oc.UseOCWithSSH(sshRunner)
		operatorsStatus, err := cluster.GetClusterOperatorsStatus(ocConfig)
		if err != nil {
			openshiftStatus = "Not Reachable"
			logging.Debug(err.Error())
		}
		switch {
		case operatorsStatus.Available:
			openshiftVersion = crcBundleMetadata.GetOpenshiftVersion()
			openshiftStatus = "Running"
		case operatorsStatus.Degraded:
			openshiftStatus = "Degraded"
		case operatorsStatus.Progressing:
			openshiftStatus = "Starting"
		}
		diskSize, diskUse, err = cluster.GetRootPartitionUsage(sshRunner)
		if err != nil {
			return statusError(statusConfig.Name, "Cannot get root partition usage", err)
		}
	}
	return ClusterStatusResult{
		Name:             statusConfig.Name,
		CrcStatus:        vmStatus.String(),
		OpenshiftStatus:  openshiftStatus,
		OpenshiftVersion: openshiftVersion,
		DiskUse:          diskUse,
		DiskSize:         diskSize,
		Success:          true,
	}, nil
}

func (*client) Exists(name string) (bool, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return false, err
	}
	exists, err := libMachineAPIClient.Exists(name)
	if err != nil {
		return false, fmt.Errorf("Error checking if the host exists: %s", err)
	}
	return exists, nil
}

func createHost(api libmachine.API, machineConfig config.MachineConfig) (*host.Host, error) {
	vm, err := newHost(api, machineConfig)
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
}

func addNameServerToInstance(sshRunner *crcssh.Runner, ns string) error {
	nameserver := network.NameServer{IPAddress: ns}
	nameservers := []network.NameServer{nameserver}
	exist, err := network.HasGivenNameserversConfigured(sshRunner, nameserver)
	if err != nil {
		return err
	}
	if !exist {
		logging.Infof("Adding %s as nameserver to the instance ...", nameserver.IPAddress)
		return network.AddNameserversToInstance(sshRunner, nameservers)
	}
	return nil
}

// Return console URL if the VM is present.
func (*client) GetConsoleURL(consoleConfig ConsoleConfig) (ConsoleResult, error) {
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return consoleURLError("Cannot initialize libmachine", err)
	}
	host, err := libMachineAPIClient.Load(consoleConfig.Name)
	if err != nil {
		return consoleURLError("Cannot load machine", err)
	}

	vmState, err := host.Driver.GetState()
	if err != nil {
		return consoleURLError("Error getting the state for host", err)
	}

	_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return consoleURLError("Error loading bundle metadata", err)
	}

	clusterConfig, err := getClusterConfig(crcBundleMetadata)
	if err != nil {
		return consoleURLError("Error loading cluster configuration", err)
	}

	return ConsoleResult{
		Success:       true,
		ClusterConfig: *clusterConfig,
		State:         vmState,
	}, nil
}

func updateSSHKeyPair(sshRunner *crcssh.Runner) error {
	if _, err := os.Stat(constants.GetPrivateKeyPath()); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// Generate ssh key pair
		logging.Info("Generating new SSH Key pair ...")
		if err := ssh.GenerateSSHKey(constants.GetPrivateKeyPath()); err != nil {
			return fmt.Errorf("Error generating ssh key pair: %v", err)
		}
	}

	// Read generated public key
	publicKey, err := ioutil.ReadFile(constants.GetPublicKeyPath())
	if err != nil {
		return err
	}

	authorizedKeys, err := sshRunner.Run("cat /home/core/.ssh/authorized_keys")
	if err == nil && strings.TrimSpace(authorizedKeys) == strings.TrimSpace(string(publicKey)) {
		return nil
	}

	logging.Info("Updating authorized keys ...")
	cmd := fmt.Sprintf("echo '%s' > /home/core/.ssh/authorized_keys; chmod 644 /home/core/.ssh/authorized_keys", publicKey)
	_, err = sshRunner.Run(cmd)
	if err != nil {
		return err
	}
	return err
}

func updateSSHKeyAndCopyKubeconfig(sshRunner *crcssh.Runner, startConfig StartConfig, crcBundleMetadata *bundle.CrcBundleInfo) error {
	if err := updateSSHKeyPair(sshRunner); err != nil {
		return fmt.Errorf("Error updating SSH Keys: %v", err)
	}

	kubeConfigFilePath := filepath.Join(constants.MachineInstanceDir, startConfig.Name, "kubeconfig")
	if _, err := os.Stat(kubeConfigFilePath); err == nil {
		return nil
	}

	// Copy Kubeconfig file from bundle extract path to machine directory.
	// In our case it would be ~/.crc/machines/crc/
	logging.Info("Copying kubeconfig file to instance dir ...")
	err := crcos.CopyFileContents(crcBundleMetadata.GetKubeConfigPath(),
		kubeConfigFilePath,
		0644)
	if err != nil {
		return fmt.Errorf("Error copying kubeconfig file to instance dir: %v", err)
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

func configProxyForCluster(ocConfig oc.Config, sshRunner *crcssh.Runner, sd *systemd.Commander, proxy *network.ProxyConfig, instanceIP string) (err error) {
	if !proxy.IsEnabled() {
		return nil
	}
	defer func() {
		// Restart the crio service
		if proxy.IsEnabled() {
			// Restart reload the daemon and then restart the service
			// So no need to explicit reload the daemon.
			if ferr := sd.Restart("crio"); ferr != nil {
				err = ferr
			}
			if ferr := sd.Restart("kubelet"); ferr != nil {
				err = ferr
			}
		}
	}()

	logging.Info("Adding proxy configuration to the cluster ...")
	proxy.AddNoProxy(instanceIP)
	if err := cluster.AddProxyConfigToCluster(sshRunner, ocConfig, proxy); err != nil {
		return err
	}

	logging.Info("Adding proxy configuration to kubelet and crio service ...")
	if err := cluster.AddProxyToKubeletAndCriO(sshRunner, proxy); err != nil {
		return err
	}

	return nil
}

func waitForProxyPropagation(ocConfig oc.Config, proxyConfig *network.ProxyConfig) {
	checkProxySettingsForOperator := func() error {
		proxySet, err := cluster.CheckProxySettingsForOperator(ocConfig, proxyConfig, "redhat-operators", "openshift-marketplace")
		if err != nil {
			logging.Debugf("Error getting proxy setting for openshift-marketplace operator %v", err)
			return &errors.RetriableError{Err: err}
		}
		if !proxySet {
			logging.Debug("Proxy changes for cluster in progress")
			return &errors.RetriableError{Err: fmt.Errorf("")}
		}
		return nil
	}

	if err := errors.RetryAfter(60*time.Second, checkProxySettingsForOperator, 2*time.Second); err != nil {
		logging.Debug("Failed to propagate proxy settings to cluster")
	}
}

func logBundleDate(crcBundleMetadata *bundle.CrcBundleInfo) {
	if buildTime, err := crcBundleMetadata.GetBundleBuildTime(); err == nil {
		bundleAgeDays := time.Since(buildTime).Hours() / 24
		if bundleAgeDays >= 30 {
			/* Initial bundle certificates are only valid for 30 days */
			logging.Debugf("Bundle has been generated %d days ago", int(bundleAgeDays))
		}
	}
}
