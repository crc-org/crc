package machine

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/ssh"
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
	"github.com/code-ready/machine/libmachine/state"
)

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
		if err == drivers.ErrNotImplemented {
			logging.Warn("Memory configuration change has been ignored as the machine driver does not support it")
		} else {
			return err
		}
	}
	if err := setVcpus(host, startConfig.CPUs); err != nil {
		logging.Debugf("Failed to update CRC VM configuration: %v", err)
		if err == drivers.ErrNotImplemented {
			logging.Warn("CPU configuration change has been ignored as the machine driver does not support it")
		} else {
			return err
		}
	}
	if err := api.Save(host); err != nil {
		return err
	}

	/* Disk size */
	if startConfig.DiskSize != constants.DefaultDiskSize {
		if err := setDiskSize(host, startConfig.DiskSize); err != nil {
			logging.Debugf("Failed to update CRC disk configuration: %v", err)
			if err == drivers.ErrNotImplemented {
				logging.Warn("Disk size configuration change has been ignored as the machine driver does not support it")
			} else {
				return err
			}
		}
		if err := api.Save(host); err != nil {
			return err
		}
	}

	return nil
}

func (client *client) Start(startConfig StartConfig) (*StartResult, error) {
	var crcBundleMetadata *bundle.CrcBundleInfo

	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot initialize libmachine")
	}

	// Pre-VM start
	var host *host.Host
	exists, err := client.Exists()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot determine if VM exists")
	}
	if !exists {
		// Ask early for pull secret if it hasn't been requested yet
		_, err = startConfig.PullSecret.Value()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to ask for pull secret")
		}

		machineConfig := config.MachineConfig{
			Name:       client.name,
			BundleName: filepath.Base(startConfig.BundlePath),
			CPUs:       startConfig.CPUs,
			Memory:     startConfig.Memory,
		}

		crcBundleMetadata, err = getCrcBundleInfo(startConfig.BundlePath)
		if err != nil {
			return nil, errors.Wrap(err, "Error getting bundle metadata")
		}

		logging.Infof("Checking size of the disk image %s ...", crcBundleMetadata.GetDiskImagePath())
		if err := crcBundleMetadata.CheckDiskImageSize(); err != nil {
			return nil, errors.Wrapf(err, "Invalid bundle disk image '%s'", crcBundleMetadata.GetDiskImagePath())
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
			return nil, errors.Wrap(err, "Error creating machine")
		}
	} else {
		host, err = libMachineAPIClient.Load(client.name)
		if err != nil {
			return nil, errors.Wrap(err, "Error loading machine")
		}

		var bundleName string
		bundleName, crcBundleMetadata, err = getBundleMetadataFromDriver(host.Driver)
		if err != nil {
			return nil, errors.Wrap(err, "Error loading bundle metadata")
		}
		if bundleName != filepath.Base(startConfig.BundlePath) {
			logging.Debugf("Bundle '%s' was requested, but the existing VM is using '%s'",
				filepath.Base(startConfig.BundlePath), bundleName)
			return nil, fmt.Errorf("Bundle '%s' was requested, but the existing VM is using '%s'",
				filepath.Base(startConfig.BundlePath),
				bundleName)
		}
		vmState, err := host.Driver.GetState()
		if err != nil {
			return nil, errors.Wrap(err, "Error getting the machine state")
		}
		if vmState == state.Running {
			logging.Infof("A CodeReady Containers VM for OpenShift %s is already running", crcBundleMetadata.GetOpenshiftVersion())
			clusterConfig, err := getClusterConfig(crcBundleMetadata)
			if err != nil {
				return nil, errors.Wrap(err, "Cannot create cluster configuration")
			}
			return &StartResult{
				Status:         vmState,
				ClusterConfig:  *clusterConfig,
				KubeletStarted: true,
			}, nil
		}

		logging.Infof("Starting CodeReady Containers VM for OpenShift %s...", crcBundleMetadata.GetOpenshiftVersion())

		if err := client.updateVMConfig(startConfig, libMachineAPIClient, host); err != nil {
			return nil, errors.Wrap(err, "Could not update CRC VM configuration")
		}

		if err := host.Driver.Start(); err != nil {
			return nil, errors.Wrap(err, "Error starting stopped VM")
		}
	}

	clusterConfig, err := getClusterConfig(crcBundleMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot create cluster configuration")
	}

	// Post-VM start
	vmState, err := host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the state")
	}
	if vmState != state.Running {
		return nil, errors.Wrap(err, "CodeReady Containers VM is not running")
	}

	instanceIP, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the IP")
	}
	sshRunner := crcssh.CreateRunner(instanceIP, constants.DefaultSSHPort, crcBundleMetadata.GetSSHKeyPath(), constants.GetPrivateKeyPath())

	logging.Debug("Waiting until ssh is available")
	if err := cluster.WaitForSSH(sshRunner); err != nil {
		return nil, errors.Wrap(err, "Failed to connect to the CRC VM with SSH -- host might be unreachable")
	}
	logging.Info("CodeReady Containers VM is running")

	// Post VM start immediately update SSH key and copy kubeconfig to instance
	// dir and VM
	if err := updateSSHKeyAndCopyKubeconfig(sshRunner, client.name, crcBundleMetadata); err != nil {
		return nil, errors.Wrap(err, "Error updating public key")
	}

	// Trigger disk resize, this will be a no-op if no disk size change is needed
	if _, err = sshRunner.Run("sudo xfs_growfs / >/dev/null"); err != nil {
		return nil, errors.Wrap(err, "Error updating filesystem size")
	}

	// Start network time synchronization if `CRC_DEBUG_ENABLE_STOP_NTP` is not set
	if stopNtp, _ := strconv.ParseBool(os.Getenv("CRC_DEBUG_ENABLE_STOP_NTP")); !stopNtp {
		logging.Info("Starting network time synchronization in CodeReady Containers VM")
		if _, err := sshRunner.Run("sudo timedatectl set-ntp on"); err != nil {
			return nil, errors.Wrap(err, "Failed to start network time synchronization")
		}
	}

	// Add nameserver to VM if provided by User
	if startConfig.NameServer != "" {
		if err = addNameServerToInstance(sshRunner, startConfig.NameServer); err != nil {
			return nil, errors.Wrap(err, "Failed to add nameserver to the VM")
		}
	}

	proxyConfig, err := getProxyConfig(crcBundleMetadata.ClusterInfo.BaseDomain)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting proxy configuration")
	}
	proxyConfig.ApplyToEnvironment()

	// Create servicePostStartConfig for DNS checks and DNS start.
	servicePostStartConfig := services.ServicePostStartConfig{
		Name: client.name,
		// TODO: would prefer passing in a more generic type
		SSHRunner: sshRunner,
		IP:        instanceIP,
		// TODO: should be more finegrained
		BundleMetadata: *crcBundleMetadata,
	}

	// Run the DNS server inside the VM
	if err := dns.RunPostStart(servicePostStartConfig); err != nil {
		return nil, errors.Wrap(err, "Error running post start")
	}

	// Check DNS lookup before starting the kubelet
	if queryOutput, err := dns.CheckCRCLocalDNSReachable(servicePostStartConfig); err != nil {
		return nil, errors.Wrapf(err, "Failed internal DNS query: %s", queryOutput)
	}
	logging.Info("Check internal and public DNS query ...")

	if queryOutput, err := dns.CheckCRCPublicDNSReachable(servicePostStartConfig); err != nil {
		logging.Warnf("Failed public DNS query from the cluster: %v : %s", err, queryOutput)
	}

	// Check DNS lookup from host to VM
	logging.Info("Check DNS query from host ...")
	if err := network.CheckCRCLocalDNSReachableFromHost(crcBundleMetadata, instanceIP); err != nil {
		return nil, errors.Wrap(err, "Failed to query DNS from host")
	}

	if err := cluster.EnsurePullSecretPresentOnInstanceDisk(sshRunner, startConfig.PullSecret); err != nil {
		return nil, errors.Wrap(err, "Failed to update VM pull secret")
	}

	// Check the certs validity inside the vm
	logging.Info("Verifying validity of the kubelet certificates ...")
	clientExpired, serverExpired, err := cluster.CheckCertsValidity(sshRunner)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to check certificate validity")
	}

	logging.Info("Starting OpenShift kubelet service")
	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	if err := sd.Start("kubelet"); err != nil {
		return nil, errors.Wrap(err, "Error starting kubelet")
	}

	ocConfig := oc.UseOCWithSSH(sshRunner)

	if err := cluster.ApproveCSRAndWaitForCertsRenewal(sshRunner, ocConfig, clientExpired, serverExpired); err != nil {
		logBundleDate(crcBundleMetadata)
		return nil, errors.Wrap(err, "Failed to renew TLS certificates: please check if a newer CodeReady Containers release is available")
	}

	if !exists {
		logging.Info("Configuring cluster for first start")
		if err := configProxyForCluster(ocConfig, sshRunner, sd, proxyConfig, instanceIP); err != nil {
			return nil, errors.Wrap(err, "Error Setting cluster config")
		}
	}

	if err := cluster.EnsurePullSecretPresentInTheCluster(ocConfig, startConfig.PullSecret); err != nil {
		return nil, errors.Wrap(err, "Failed to update cluster pull secret")
	}

	if err := cluster.EnsureClusterIDIsNotEmpty(ocConfig); err != nil {
		return nil, errors.Wrap(err, "Failed to update cluster ID")
	}

	// Check if kubelet service is running inside the VM
	kubeletStatus, err := sd.Status("kubelet")
	if err != nil || kubeletStatus != states.Running {
		return nil, errors.Wrap(err, "kubelet service is not running")
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
		return nil, errors.Wrap(err, "Failed to wait for the client-ca request header update")
	}

	if err := cluster.DeleteOpenshiftAPIServerPods(ocConfig); err != nil {
		return nil, errors.Wrap(err, "Cannot delete OpenShift API Server pods")
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
	return &StartResult{
		KubeletStarted: true,
		ClusterConfig:  *clusterConfig,
		Status:         vmState,
	}, nil
}

func (client *client) Stop() (state.State, error) {
	defer unsetMachineLogging()

	// Set libmachine logging
	err := setMachineLogging(client.debug)
	if err != nil {
		return state.None, errors.Wrap(err, "Cannot initialize logging")
	}

	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return state.None, errors.Wrap(err, "Cannot initialize libmachine")
	}
	host, err := libMachineAPIClient.Load(client.name)

	if err != nil {
		return state.None, errors.Wrap(err, "Cannot load machine")
	}

	// FIXME: Why is the state fetched before calling host.Stop() ? We will return state.Running most of the time instead of state.Stopped
	vmState, _ := host.Driver.GetState()
	logging.Info("Stopping the OpenShift cluster, this may take a few minutes...")
	if err := host.Stop(); err != nil {
		return state.None, errors.Wrap(err, "Cannot stop machine")
	}
	return vmState, nil
}

func (client *client) PowerOff() error {
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return errors.Wrap(err, "Cannot initialize libmachine")
	}

	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return errors.Wrap(err, "Cannot load machine")
	}

	if err := host.Kill(); err != nil {
		return errors.Wrap(err, "Cannot kill machine")
	}
	return nil
}

func (client *client) Delete() error {
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return errors.Wrap(err, "Cannot initialize libmachine")
	}
	host, err := libMachineAPIClient.Load(client.name)

	if err != nil {
		return errors.Wrap(err, "Cannot load machine")
	}

	if err := host.Driver.Remove(); err != nil {
		return errors.Wrap(err, "Driver cannot remove machine")
	}

	if err := libMachineAPIClient.Remove(client.name); err != nil {
		return errors.Wrap(err, "Cannot remove machine")
	}
	return nil
}

func (client *client) IP() (string, error) {
	err := setMachineLogging(client.debug)
	if err != nil {
		return "", errors.Wrap(err, "Cannot initialize logging")
	}

	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return "", errors.Wrap(err, "Cannot initialize libmachine")
	}
	host, err := libMachineAPIClient.Load(client.name)

	if err != nil {
		return "", errors.Wrap(err, "Cannot load machine")
	}
	ip, err := host.Driver.GetIP()
	if err != nil {
		return "", errors.Wrap(err, "Cannot get IP")
	}
	return ip, nil
}

func (client *client) Status() (*ClusterStatusResult, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(false)
	defer cleanup()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot initialize libmachine")
	}

	_, err = libMachineAPIClient.Exists(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot check if machine exists")
	}

	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}
	vmStatus, err := host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get machine state")
	}

	if vmStatus != state.Running {
		return &ClusterStatusResult{
			CrcStatus:       vmStatus,
			OpenshiftStatus: "Stopped",
		}, nil
	}

	_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bundle metadata")
	}
	proxyConfig, err := getProxyConfig(crcBundleMetadata.ClusterInfo.BaseDomain)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting proxy configuration")
	}
	proxyConfig.ApplyToEnvironment()

	ip, err := host.Driver.GetIP()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting ip")
	}
	sshRunner := crcssh.CreateRunner(ip, constants.DefaultSSHPort, constants.GetPrivateKeyPath())
	// check if all the clusteroperators are running
	diskSize, diskUse, err := cluster.GetRootPartitionUsage(sshRunner)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get root partition usage")
	}
	return &ClusterStatusResult{
		CrcStatus:        state.Running,
		OpenshiftStatus:  getOpenShiftStatus(sshRunner),
		OpenshiftVersion: crcBundleMetadata.GetOpenshiftVersion(),
		DiskUse:          diskUse,
		DiskSize:         diskSize,
	}, nil
}

func getOpenShiftStatus(sshRunner *crcssh.Runner) string {
	status, err := cluster.GetClusterOperatorsStatus(oc.UseOCWithSSH(sshRunner))
	if err != nil {
		logging.Debugf("cannot get OpenShift status: %v", err)
		return "Not Reachable"
	}
	if status.Progressing {
		return "Starting"
	}
	if status.Degraded {
		return "Degraded"
	}
	if status.Available {
		return "Running"
	}
	return "Stopped"
}

func (client *client) Exists() (bool, error) {
	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return false, err
	}
	exists, err := libMachineAPIClient.Exists(client.name)
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
func (client *client) GetConsoleURL() (*ConsoleResult, error) {
	// Here we are only checking if the VM exist and not the status of the VM.
	// We might need to improve and use crc status logic, only
	// return if the Openshift is running as part of status.
	libMachineAPIClient, cleanup, err := createLibMachineClient(client.debug)
	defer cleanup()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot initialize libmachine")
	}
	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load machine")
	}

	vmState, err := host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the state for host")
	}

	_, crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bundle metadata")
	}

	clusterConfig, err := getClusterConfig(crcBundleMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading cluster configuration")
	}

	return &ConsoleResult{
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

func updateSSHKeyAndCopyKubeconfig(sshRunner *crcssh.Runner, name string, crcBundleMetadata *bundle.CrcBundleInfo) error {
	if err := updateSSHKeyPair(sshRunner); err != nil {
		return fmt.Errorf("Error updating SSH Keys: %v", err)
	}

	kubeConfigFilePath := filepath.Join(constants.MachineInstanceDir, name, "kubeconfig")
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
			return &crcerrors.RetriableError{Err: err}
		}
		if !proxySet {
			logging.Debug("Proxy changes for cluster in progress")
			return &crcerrors.RetriableError{Err: fmt.Errorf("")}
		}
		return nil
	}

	if err := crcerrors.RetryAfter(60*time.Second, checkProxySettingsForOperator, 2*time.Second); err != nil {
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
