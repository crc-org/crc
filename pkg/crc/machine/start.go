package machine

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/code-ready/crc/pkg/crc/cluster"
	"github.com/code-ready/crc/pkg/crc/constants"
	crcerrors "github.com/code-ready/crc/pkg/crc/errors"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/machine/bundle"
	"github.com/code-ready/crc/pkg/crc/machine/config"
	"github.com/code-ready/crc/pkg/crc/machine/state"
	"github.com/code-ready/crc/pkg/crc/machine/types"
	"github.com/code-ready/crc/pkg/crc/network"
	"github.com/code-ready/crc/pkg/crc/oc"
	"github.com/code-ready/crc/pkg/crc/services"
	"github.com/code-ready/crc/pkg/crc/services/dns"
	crcssh "github.com/code-ready/crc/pkg/crc/ssh"
	"github.com/code-ready/crc/pkg/crc/systemd"
	"github.com/code-ready/crc/pkg/crc/telemetry"
	crctls "github.com/code-ready/crc/pkg/crc/tls"
	"github.com/code-ready/crc/pkg/libmachine"
	"github.com/code-ready/crc/pkg/libmachine/host"
	"github.com/code-ready/machine/libmachine/drivers"
	libmachinestate "github.com/code-ready/machine/libmachine/state"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const minimumMemoryForMonitoring = 14336

func getCrcBundleInfo(bundleName, bundlePath string) (*bundle.CrcBundleInfo, error) {
	bundleInfo, err := bundle.Use(bundleName)
	if err == nil {
		logging.Infof("Loading bundle: %s...", bundleName)
		return bundleInfo, nil
	}
	logging.Debugf("Failed to load bundle %s: %v", bundleName, err)
	logging.Infof("Extracting bundle: %s...", bundleName)
	if _, err := bundle.Extract(bundlePath); err != nil {
		return nil, err
	}
	return bundle.Use(bundleName)
}

func (client *client) updateVMConfig(startConfig types.StartConfig, api libmachine.API, host *host.Host) error {
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

func growRootFileSystem(sshRunner *crcssh.Runner) error {
	// With 4.7, this is quite a manual process until https://github.com/openshift/installer/pull/4746 gets fixed
	// See https://github.com/code-ready/crc/issues/2104 for details
	rootPart, _, err := sshRunner.Run("realpath", "/dev/disk/by-label/root")
	if err != nil {
		return err
	}
	rootPart = strings.TrimSpace(rootPart)
	if !strings.HasPrefix(rootPart, "/dev/vda") && !strings.HasPrefix(rootPart, "/dev/sda") {
		return fmt.Errorf("Unexpected root device: %s", rootPart)
	}
	// with '/dev/[sv]da4' as input, run 'growpart /dev/[sv]da 4'
	if _, _, err := sshRunner.RunPrivileged(fmt.Sprintf("Growing %s partition", rootPart), "/usr/bin/growpart", rootPart[:len("/dev/.da")], rootPart[len(rootPart)-1:]); err != nil {
		var exitErr *ssh.ExitError
		if !errors.As(err, &exitErr) {
			return err
		}
		if exitErr.ExitStatus() != 1 {
			return err
		}

		logging.Debugf("No free space after %s, nothing to do", rootPart)
		return nil
	}

	logging.Infof("Resizing %s filesystem", rootPart)
	if _, _, err := sshRunner.RunPrivileged("Remounting /sysroot read/write", "mount -o remount,rw /sysroot"); err != nil {
		return err
	}
	if _, _, err = sshRunner.RunPrivileged(fmt.Sprintf("Growing %s filesystem", rootPart), "xfs_growfs", rootPart); err != nil {
		return err
	}

	return nil
}
func (client *client) Start(ctx context.Context, startConfig types.StartConfig) (*types.StartResult, error) {
	telemetry.SetCPUs(ctx, startConfig.CPUs)
	telemetry.SetMemory(ctx, uint64(startConfig.Memory)*1024*1024)
	telemetry.SetDiskSize(ctx, uint64(startConfig.DiskSize)*1024*1024*1024)

	if err := client.validateStartConfig(startConfig); err != nil {
		return nil, err
	}

	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()

	// Pre-VM start
	exists, err := client.Exists()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot determine if VM exists")
	}

	bundleName := bundle.GetBundleNameWithoutExtension(filepath.Base(startConfig.BundlePath))

	if !exists {
		telemetry.SetStartType(ctx, telemetry.CreationStartType)

		// Ask early for pull secret if it hasn't been requested yet
		_, err = startConfig.PullSecret.Value()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to ask for pull secret")
		}

		crcBundleMetadata, err := getCrcBundleInfo(bundleName, startConfig.BundlePath)
		if err != nil {
			return nil, errors.Wrap(err, "Error getting bundle metadata")
		}

		logging.Infof("Creating CodeReady Containers VM for OpenShift %s...", crcBundleMetadata.GetOpenshiftVersion())

		machineConfig := config.MachineConfig{
			Name:            client.name,
			BundleName:      bundleName,
			CPUs:            startConfig.CPUs,
			Memory:          startConfig.Memory,
			DiskSize:        startConfig.DiskSize,
			NetworkMode:     client.networkMode(),
			ImageSourcePath: crcBundleMetadata.GetDiskImagePath(),
			ImageFormat:     crcBundleMetadata.GetDiskImageFormat(),
			SSHKeyPath:      crcBundleMetadata.GetSSHKeyPath(),
			KernelCmdLine:   crcBundleMetadata.GetKernelCommandLine(),
			Initramfs:       crcBundleMetadata.GetInitramfsPath(),
			Kernel:          crcBundleMetadata.GetKernelPath(),
			KubeConfig:      crcBundleMetadata.GetKubeConfigPath(),
		}
		if err := createHost(libMachineAPIClient, machineConfig); err != nil {
			return nil, errors.Wrap(err, "Error creating machine")
		}
	} else {
		telemetry.SetStartType(ctx, telemetry.StartStartType)
	}

	host, err := libMachineAPIClient.Load(client.name)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading machine")
	}

	crcBundleMetadata, err := getBundleMetadataFromDriver(host.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading bundle metadata")
	}
	currentBundleName := crcBundleMetadata.GetBundleName()
	if currentBundleName != bundleName {
		logging.Debugf("Bundle '%s' was requested, but the existing VM is using '%s'",
			bundleName, currentBundleName)
		return nil, fmt.Errorf("Bundle '%s' was requested, but the existing VM is using '%s'. Please delete your existing cluster and start again",
			bundleName,
			currentBundleName)
	}
	vmState, err := host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the machine state")
	}
	if vmState == libmachinestate.Running {
		logging.Infof("A CodeReady Containers VM for OpenShift %s is already running", crcBundleMetadata.GetOpenshiftVersion())
		clusterConfig, err := getClusterConfig(crcBundleMetadata)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot create cluster configuration")
		}

		telemetry.SetStartType(ctx, telemetry.AlreadyRunningStartType)
		return &types.StartResult{
			Status:         state.FromMachine(vmState),
			ClusterConfig:  *clusterConfig,
			KubeletStarted: true,
		}, nil
	}

	if _, err := bundle.Use(currentBundleName); err != nil {
		return nil, err
	}

	logging.Infof("Starting CodeReady Containers VM for OpenShift %s...", crcBundleMetadata.GetOpenshiftVersion())

	if client.useVSock() {
		if err := exposePorts(); err != nil {
			return nil, err
		}
	}

	if err := client.updateVMConfig(startConfig, libMachineAPIClient, host); err != nil {
		return nil, errors.Wrap(err, "Could not update CRC VM configuration")
	}

	if err := startHost(ctx, libMachineAPIClient, host); err != nil {
		return nil, errors.Wrap(err, "Error starting machine")
	}

	// Post-VM start
	vmState, err = host.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the state")
	}
	if vmState != libmachinestate.Running {
		return nil, errors.Wrap(err, "CodeReady Containers VM is not running")
	}

	instanceIP, err := getIP(host, client.useVSock())
	if err != nil {
		return nil, errors.Wrap(err, "Error getting the IP")
	}
	logging.Infof("CodeReady Containers instance is running with IP %s", instanceIP)
	sshRunner, err := crcssh.CreateRunner(instanceIP, getSSHPort(client.useVSock()), crcBundleMetadata.GetSSHKeyPath(), constants.GetPrivateKeyPath(), constants.GetRsaPrivateKeyPath())
	if err != nil {
		return nil, errors.Wrap(err, "Error creating the ssh client")
	}
	defer sshRunner.Close()

	logging.Debug("Waiting until ssh is available")
	if err := sshRunner.WaitForConnectivity(ctx, 300*time.Second); err != nil {
		return nil, errors.Wrap(err, "Failed to connect to the CRC VM with SSH -- host might be unreachable")
	}
	logging.Info("CodeReady Containers VM is running")

	// Post VM start immediately update SSH key and copy kubeconfig to instance
	// dir and VM
	if err := updateSSHKeyPair(sshRunner); err != nil {
		return nil, errors.Wrap(err, "Error updating public key")
	}

	// Trigger disk resize, this will be a no-op if no disk size change is needed
	if err := growRootFileSystem(sshRunner); err != nil {
		return nil, errors.Wrap(err, "Error updating filesystem size")
	}

	// Start network time synchronization if `CRC_DEBUG_ENABLE_STOP_NTP` is not set
	if stopNtp, _ := strconv.ParseBool(os.Getenv("CRC_DEBUG_ENABLE_STOP_NTP")); stopNtp {
		logging.Info("Stopping network time synchronization in CodeReady Containers VM")
		if _, _, err := sshRunner.RunPrivileged("Turning off the ntp server", "timedatectl set-ntp off"); err != nil {
			return nil, errors.Wrap(err, "Failed to stop network time synchronization")
		}
		logging.Info("Setting clock to host clock (UTC timezone)")
		dateCmd := fmt.Sprintf("date -s '%s'", time.Now().Format(time.UnixDate))
		if _, _, err := sshRunner.RunPrivileged("Setting clock same as host", dateCmd); err != nil {
			return nil, errors.Wrap(err, "Failed to set clock to same as host")
		}
	}

	// Add nameserver to VM if provided by User
	if startConfig.NameServer != "" {
		if err = addNameServerToInstance(sshRunner, startConfig.NameServer); err != nil {
			return nil, errors.Wrap(err, "Failed to add nameserver to the VM")
		}
	}

	if _, _, err := sshRunner.RunPrivileged("make root Podman socket accessible", "chmod 777 /run/podman/ /run/podman/podman.sock"); err != nil {
		return nil, errors.Wrap(err, "Failed to change permissions to root podman socket")
	}

	proxyConfig, err := getProxyConfig(crcBundleMetadata.ClusterInfo.BaseDomain)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting proxy configuration")
	}
	proxyConfig.ApplyToEnvironment()
	proxyConfig.AddNoProxy(instanceIP)

	// Create servicePostStartConfig for DNS checks and DNS start.
	servicePostStartConfig := services.ServicePostStartConfig{
		Name: client.name,
		// TODO: would prefer passing in a more generic type
		SSHRunner: sshRunner,
		IP:        instanceIP,
		// TODO: should be more finegrained
		BundleMetadata: *crcBundleMetadata,
		NetworkMode:    client.networkMode(),
	}

	// Run the DNS server inside the VM
	if err := dns.RunPostStart(servicePostStartConfig); err != nil {
		return nil, errors.Wrap(err, "Error running post start")
	}

	// Check DNS lookup before starting the kubelet
	if queryOutput, err := dns.CheckCRCLocalDNSReachable(ctx, servicePostStartConfig); err != nil {
		if !client.useVSock() {
			return nil, errors.Wrapf(err, "Failed internal DNS query: %s", queryOutput)
		}
	}
	logging.Info("Check internal and public DNS query...")

	if queryOutput, err := dns.CheckCRCPublicDNSReachable(servicePostStartConfig); err != nil {
		logging.Warnf("Failed public DNS query from the cluster: %v : %s", err, queryOutput)
	}

	// Check DNS lookup from host to VM
	logging.Info("Check DNS query from host...")
	if err := network.CheckCRCLocalDNSReachableFromHost(crcBundleMetadata.GetAPIHostname(),
		crcBundleMetadata.GetAppHostname("foo"), crcBundleMetadata.ClusterInfo.AppsDomain, instanceIP); err != nil {
		if !client.useVSock() {
			return nil, errors.Wrap(err, "Failed to query DNS from host")
		}
		logging.Warn(fmt.Sprintf("Failed to query DNS from host: %v", err))
	}

	// Remove this block after 2-3 release (after v1.32.0)
	// This is just to support 4.7 bundle with current master
	if strings.HasPrefix(crcBundleMetadata.GetOpenshiftVersion(), "4.7.") {
		if err := cluster.EnsurePullSecretPresentOnInstanceDisk(sshRunner, startConfig.PullSecret); err != nil {
			return nil, errors.Wrap(err, "Failed to update VM pull secret")
		}
	}

	// Check the certs validity inside the vm
	logging.Info("Verifying validity of the kubelet certificates...")
	certsExpired, err := cluster.CheckCertsValidity(sshRunner)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to check certificate validity")
	}

	logging.Info("Starting OpenShift kubelet service")
	sd := systemd.NewInstanceSystemdCommander(sshRunner)
	if err := sd.Start("kubelet"); err != nil {
		return nil, errors.Wrap(err, "Error starting kubelet")
	}

	ocConfig := oc.UseOCWithSSH(sshRunner)

	if err := cluster.ApproveCSRAndWaitForCertsRenewal(ctx, sshRunner, ocConfig, certsExpired[cluster.KubeletClientCert], certsExpired[cluster.KubeletServerCert]); err != nil {
		logBundleDate(crcBundleMetadata)
		return nil, errors.Wrap(err, "Failed to renew TLS certificates: please check if a newer CodeReady Containers release is available")
	}

	if err := cluster.WaitForAPIServer(ctx, ocConfig); err != nil {
		return nil, errors.Wrap(err, "Error waiting for apiserver")
	}

	if err := cluster.DeleteMCOLeaderLease(ctx, ocConfig); err != nil {
		return nil, err
	}

	if err := cluster.EnsurePullSecretPresentInTheCluster(ctx, ocConfig, startConfig.PullSecret); err != nil {
		return nil, errors.Wrap(err, "Failed to update cluster pull secret")
	}

	if err := cluster.EnsureSSHKeyPresentInTheCluster(ctx, ocConfig, constants.GetPublicKeyPath()); err != nil {
		return nil, errors.Wrap(err, "Failed to update ssh public key to machine config")
	}

	if err := cluster.WaitForPullSecretPresentOnInstanceDisk(ctx, sshRunner); err != nil {
		return nil, errors.Wrap(err, "Failed to update pull secret on the disk")
	}

	if err := ensureProxyIsConfiguredInOpenShift(ctx, ocConfig, sshRunner, proxyConfig, instanceIP); err != nil {
		return nil, errors.Wrap(err, "Failed to update cluster proxy configuration")
	}

	if err := cluster.UpdateKubeAdminUserPassword(ctx, ocConfig, startConfig.KubeAdminPassword); err != nil {
		return nil, errors.Wrap(err, "Failed to update kubeadmin user password")
	}

	if err := cluster.EnsureClusterIDIsNotEmpty(ctx, ocConfig); err != nil {
		return nil, errors.Wrap(err, "Failed to update cluster ID")
	}

	if client.useVSock() {
		if err := ensureRoutesControllerIsRunning(sshRunner, ocConfig); err != nil {
			return nil, err
		}
	}

	if client.monitoringEnabled() {
		logging.Info("Enabling cluster monitoring operator...")
		if err := cluster.StartMonitoring(ocConfig); err != nil {
			return nil, errors.Wrap(err, "Cannot start monitoring stack")
		}
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
	if certsExpired[cluster.AggregatorClientCert] {
		logging.Debug("Waiting for the renewal of the request header client ca...")
		if err := cluster.WaitForRequestHeaderClientCaFile(ctx, sshRunner); err != nil {
			return nil, errors.Wrap(err, "Failed to wait for aggregator client ca renewal")
		}

		if err := cluster.DeleteOpenshiftAPIServerPods(ctx, ocConfig); err != nil {
			return nil, errors.Wrap(err, "Cannot delete OpenShift API Server pods")
		}
	}

	if err := updateKubeconfig(ctx, ocConfig, sshRunner, crcBundleMetadata.GetKubeConfigPath()); err != nil {
		return nil, errors.Wrap(err, "Failed to update kubeconfig file")
	}

	logging.Info("Starting OpenShift cluster... [waiting for the cluster to stabilize]")
	if err := cluster.WaitForClusterStable(ctx, instanceIP, constants.KubeconfigFilePath, proxyConfig); err != nil {
		logging.Errorf("Cluster is not ready: %v", err)
	}

	waitForProxyPropagation(ctx, ocConfig, proxyConfig)

	clusterConfig, err := getClusterConfig(crcBundleMetadata)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get cluster configuration")
	}

	logging.Info("Adding crc-admin and crc-developer contexts to kubeconfig...")
	if err := writeKubeconfig(instanceIP, clusterConfig); err != nil {
		logging.Errorf("Cannot update kubeconfig: %v", err)
	}

	return &types.StartResult{
		KubeletStarted: true,
		ClusterConfig:  *clusterConfig,
		Status:         state.FromMachine(vmState),
	}, nil
}

func (client *client) IsRunning() (bool, error) {
	libMachineAPIClient, cleanup := createLibMachineClient()
	defer cleanup()
	host, err := libMachineAPIClient.Load(client.name)

	if err != nil {
		return false, errors.Wrap(err, "Cannot load machine")
	}

	// get the actual state
	vmState, err := host.Driver.GetState()
	if err != nil {
		// but reports not started on error
		return false, errors.Wrap(err, "Error getting the state")
	}
	if vmState != libmachinestate.Running {
		return false, nil
	}
	return true, nil
}

func (client *client) validateStartConfig(startConfig types.StartConfig) error {
	if client.monitoringEnabled() && startConfig.Memory < minimumMemoryForMonitoring {
		return fmt.Errorf("Too little memory (%s) allocated to the virtual machine to start the monitoring stack, %s is the minimum",
			units.BytesSize(float64(startConfig.Memory)*1024*1024),
			units.BytesSize(minimumMemoryForMonitoring*1024*1024))
	}
	return nil
}

func createHost(api libmachine.API, machineConfig config.MachineConfig) error {
	vm, err := newHost(api, machineConfig)
	if err != nil {
		return fmt.Errorf("Error creating new host: %s", err)
	}

	logging.Debug("Running pre-create checks...")

	if err := vm.Driver.PreCreateCheck(); err != nil {
		return errors.Wrap(err, "error with pre-create check")
	}

	if err := api.Save(vm); err != nil {
		return fmt.Errorf("Error saving host to store before attempting creation: %s", err)
	}

	logging.Debug("Creating machine...")

	if err := vm.Driver.Create(); err != nil {
		return fmt.Errorf("Error in driver during machine creation: %s", err)
	}

	logging.Info("Generating new SSH Key pair...")
	if err := crcssh.GenerateSSHKey(constants.GetPrivateKeyPath()); err != nil {
		return fmt.Errorf("Error generating ssh key pair: %v", err)
	}
	if err := cluster.GenerateKubeAdminUserPassword(); err != nil {
		return errors.Wrap(err, "Error generating new kubeadmin password")
	}
	if err := api.SetExists(vm.Name); err != nil {
		return fmt.Errorf("Failed to record VM existence: %s", err)
	}

	logging.Debug("Machine successfully created")
	return nil
}

func startHost(ctx context.Context, api libmachine.API, vm *host.Host) error {
	if err := vm.Driver.Start(); err != nil {
		return fmt.Errorf("Error in driver during machine start: %s", err)
	}

	if err := api.Save(vm); err != nil {
		return fmt.Errorf("Error saving host to store after attempting creation: %s", err)
	}

	logging.Debug("Waiting for machine to be running, this may take a few minutes...")
	if err := crcerrors.Retry(ctx, 3*time.Minute, host.MachineInState(vm.Driver, libmachinestate.Running), 3*time.Second); err != nil {
		return fmt.Errorf("Error waiting for machine to be running: %s", err)
	}

	logging.Debug("Machine is up and running!")

	return nil
}

func addNameServerToInstance(sshRunner *crcssh.Runner, ns string) error {
	nameserver := network.NameServer{IPAddress: ns}
	nameservers := []network.NameServer{nameserver}
	exist, err := network.HasGivenNameserversConfigured(sshRunner, nameserver)
	if err != nil {
		return err
	}
	if !exist {
		logging.Infof("Adding %s as nameserver to the instance...", nameserver.IPAddress)
		return network.AddNameserversToInstance(sshRunner, nameservers)
	}
	return nil
}

func updateSSHKeyPair(sshRunner *crcssh.Runner) error {
	// Read generated public key
	publicKey, err := ioutil.ReadFile(constants.GetPublicKeyPath())
	if err != nil {
		return err
	}

	authorizedKeys, _, err := sshRunner.Run("cat /home/core/.ssh/authorized_keys")
	if err == nil && strings.TrimSpace(authorizedKeys) == strings.TrimSpace(string(publicKey)) {
		return nil
	}

	logging.Info("Updating authorized keys...")
	// CopyData uses sudo and we need to use it
	// because of https://bugzilla.redhat.com/show_bug.cgi?id=1956739
	err = sshRunner.CopyData(publicKey, "/home/core/.ssh/authorized_keys", 0644)
	if err != nil {
		return err
	}
	return err
}

func copyKubeconfigFileWithUpdatedUserClientCertAndKey(selfSignedCAKey *rsa.PrivateKey, selfSignedCACert *x509.Certificate, srcKubeConfigPath, dstKubeConfigPath string) error {
	if _, err := os.Stat(constants.KubeconfigFilePath); err == nil {
		return nil
	}
	clientKey, clientCert, err := crctls.GenerateClientCertificate(selfSignedCAKey, selfSignedCACert)
	if err != nil {
		return err
	}
	return updateClientCrtAndKeyToKubeconfig(clientKey, clientCert, srcKubeConfigPath, dstKubeConfigPath)
}

func ensureProxyIsConfiguredInOpenShift(ctx context.Context, ocConfig oc.Config, sshRunner *crcssh.Runner, proxy *network.ProxyConfig, instanceIP string) (err error) {
	if !proxy.IsEnabled() {
		return nil
	}
	logging.Info("Adding proxy configuration to the cluster...")
	return cluster.AddProxyConfigToCluster(ctx, sshRunner, ocConfig, proxy)
}

func waitForProxyPropagation(ctx context.Context, ocConfig oc.Config, proxyConfig *network.ProxyConfig) {
	if !proxyConfig.IsEnabled() {
		return
	}
	logging.Info("Waiting for the proxy configuration to be applied...")
	checkProxySettingsForOperator := func() error {
		proxySet, err := cluster.CheckProxySettingsForOperator(ocConfig, proxyConfig, "marketplace-operator", "openshift-marketplace")
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

	if err := crcerrors.Retry(ctx, 300*time.Second, checkProxySettingsForOperator, 2*time.Second); err != nil {
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

func ensureRoutesControllerIsRunning(sshRunner *crcssh.Runner, ocConfig oc.Config) error {
	bin, err := json.Marshal(v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "routes-controller",
			Namespace: "openshift-ingress",
		},
		Spec: v1.PodSpec{
			ServiceAccountName: "router",
			Containers: []v1.Container{
				{
					Name:            "routes-controller",
					Image:           "quay.io/crcont/routes-controller:latest",
					ImagePullPolicy: v1.PullIfNotPresent,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	if err := sshRunner.CopyData(bin, "/tmp/routes-controller.json", 0444); err != nil {
		return err
	}
	_, _, err = ocConfig.RunOcCommand("apply", "-f", "/tmp/routes-controller.json")
	if err != nil {
		return err
	}
	return nil
}

func updateKubeconfig(ctx context.Context, ocConfig oc.Config, sshRunner *crcssh.Runner, kubeconfigFilePath string) error {
	selfSignedCAKey, selfSignedCACert, err := crctls.GetSelfSignedCA()
	if err != nil {
		return errors.Wrap(err, "Not able to generate root CA key and Cert")
	}
	if err := copyKubeconfigFileWithUpdatedUserClientCertAndKey(selfSignedCAKey, selfSignedCACert, kubeconfigFilePath, constants.KubeconfigFilePath); err != nil {
		return errors.Wrapf(err, "Failed to copy kubeconfig file: %s", constants.KubeconfigFilePath)
	}
	adminClientCA, err := adminClientCertificate(constants.KubeconfigFilePath)
	if err != nil {
		return errors.Wrap(err, "Not able to get user CA")
	}
	if err := cluster.EnsureGeneratedClientCAPresentInTheCluster(ctx, ocConfig, sshRunner, selfSignedCACert, adminClientCA); err != nil {
		return errors.Wrap(err, "Failed to update user CA to cluster")
	}
	return nil
}
